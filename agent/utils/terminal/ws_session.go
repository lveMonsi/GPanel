package terminal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

const (
	// maxBufferSize 最大缓冲区大小（1MB），防止内存耗尽
	maxBufferSize = 1024 * 1024
)

// safeBuffer 线程安全的缓冲区
type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查缓冲区大小，防止无限增长
	if w.buffer.Len()+len(p) > maxBufferSize {
		// 缓冲区已满，丢弃新数据并记录日志
		return len(p), nil
	}
	return w.buffer.Write(p)
}

func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}

func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

// Len 返回缓冲区当前大小
func (w *safeBuffer) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Len()
}

const (
	// WsMsgCmd 命令消息类型
	WsMsgCmd = "cmd"
	// WsMsgResize 调整大小消息类型
	WsMsgResize = "resize"
	// WsMsgHeartbeat 心跳消息类型
	WsMsgHeartbeat = "heartbeat"
	// WsMsgConnect 连接消息类型（用于传输 SSH 凭证）
	WsMsgConnect = "connect"
)

// WsMsg WebSocket 消息
type WsMsg struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`      // WsMsgCmd
	Cols      int    `json:"cols,omitempty"`      // WsMsgResize
	Rows      int    `json:"rows,omitempty"`      // WsMsgResize
	Timestamp int    `json:"timestamp,omitempty"` // WsMsgHeartbeat
}

// LogicSshWsSession SSH WebSocket 会话
type LogicSshWsSession struct {
	stdinPipe       io.WriteCloser
	comboOutput     *safeBuffer
	logBuff         *safeBuffer
	inputFilterBuff *safeBuffer
	session         *ssh.Session
	wsConn          *websocket.Conn
	isAdmin         bool
	IsFlagged       bool
}

// NewLogicSshWsSession 创建新的 SSH WebSocket 会话
func NewLogicSshWsSession(cols, rows int, sshClient *ssh.Client, wsConn *websocket.Conn, initCmd string) (*LogicSshWsSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(safeBuffer)
	logBuf := new(safeBuffer)
	inputBuf := new(safeBuffer)
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}

	if err := sshSession.Shell(); err != nil {
		return nil, err
	}

	if len(initCmd) != 0 {
		time.Sleep(100 * time.Millisecond)
		_, _ = stdinP.Write([]byte(initCmd + "\n"))
	}

	return &LogicSshWsSession{
		stdinPipe:       stdinP,
		comboOutput:     comboWriter,
		logBuff:         logBuf,
		inputFilterBuff: inputBuf,
		session:         sshSession,
		wsConn:          wsConn,
		isAdmin:         true,
		IsFlagged:       false,
	}, nil
}

// Close 关闭会话
func (sws *LogicSshWsSession) Close() {
	if sws.stdinPipe != nil {
		_ = sws.stdinPipe.Close()
		sws.stdinPipe = nil
	}
	if sws.session != nil {
		_ = sws.session.Close()
		sws.session = nil
	}
	if sws.logBuff != nil {
		sws.logBuff = nil
	}
	if sws.comboOutput != nil {
		sws.comboOutput = nil
	}
}

// Start 启动会话
func (sws *LogicSshWsSession) Start(quitChan chan bool) {
	go sws.receiveWsMsg(quitChan)
	go sws.sendComboOutput(quitChan)
}

// receiveWsMsg 接收 WebSocket 消息
func (sws *LogicSshWsSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[A panic occurred during receive ws message, error message: %v", r)
		}
	}()

	wsConn := sws.wsConn
	defer setQuit(exitCh)

	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				return
			}

			msgObj := WsMsg{}
			_ = json.Unmarshal(wsData, &msgObj)

			switch msgObj.Type {
			case WsMsgResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.session.WindowChange(msgObj.Rows, msgObj.Cols); err != nil {
						log.Printf("ssh pty change windows size failed, err: %v", err)
					}
				}
			case WsMsgCmd:
				decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
				if err != nil {
					log.Printf("websocket cmd string base64 decoding failed, err: %v", err)
				}
				sws.sendWebsocketInputCommandToSshSessionStdinPipe(decodeBytes)
			case WsMsgHeartbeat:
				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				if err != nil {
					log.Printf("ssh sending heartbeat to websocket failed, err: %v", err)
				}
			}
		}
	}
}

// sendWebsocketInputCommandToSshSessionStdinPipe 发送命令到 SSH 会话
func (sws *LogicSshWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		log.Printf("ws cmd bytes write to ssh.stdin pipe failed, err: %v", err)
	}
}

// sendComboOutput 发送组合输出到 WebSocket
func (sws *LogicSshWsSession) sendComboOutput(exitCh chan bool) {
	wsConn := sws.wsConn
	defer setQuit(exitCh)

	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if sws.comboOutput == nil {
				return
			}

			bs := sws.comboOutput.Bytes()
			if len(bs) > 0 {
				wsData, err := json.Marshal(WsMsg{
					Type: WsMsgCmd,
					Data: base64.StdEncoding.EncodeToString(bs),
				})
				if err != nil {
					log.Printf("encoding combo output to json failed, err: %v", err)
					continue
				}

				err = wsConn.WriteMessage(websocket.TextMessage, wsData)
				if err != nil {
					log.Printf("ssh sending combo output to websocket failed, err: %v", err)
				}

				_, err = sws.logBuff.Write(bs)
				if err != nil {
					log.Printf("combo output to log buffer failed, err: %v", err)
				}

				sws.comboOutput.buffer.Reset()
			}

			// 检测 logout 命令
			if string(bs) == string([]byte{13, 10, 108, 111, 103, 111, 117, 116, 13, 10}) {
				sws.Close()
				return
			}

		case <-exitCh:
			return
		}
	}
}

// Wait 等待会话结束
func (sws *LogicSshWsSession) Wait(quitChan chan bool) {
	if err := sws.session.Wait(); err != nil {
		setQuit(quitChan)
	}
}
