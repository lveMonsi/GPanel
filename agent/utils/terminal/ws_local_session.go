package terminal

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// LocalWsSession 本地 WebSocket 会话
type LocalWsSession struct {
	slave  *LocalCommand
	wsConn *websocket.Conn

	allowCtrlC bool
	writeMutex sync.Mutex
}

// NewLocalWsSession 创建新的本地 WebSocket 会话
func NewLocalWsSession(cols, rows int, wsConn *websocket.Conn, slave *LocalCommand, allowCtrlC bool) (*LocalWsSession, error) {
	if err := slave.ResizeTerminal(cols, rows); err != nil {
		log.Printf("local pty change windows size failed, err: %v", err)
	}

	return &LocalWsSession{
		slave:  slave,
		wsConn: wsConn,

		allowCtrlC: allowCtrlC,
	}, nil
}

// Start 启动会话
func (sws *LocalWsSession) Start(quitChan chan bool) {
	go sws.handleSlaveEvent(quitChan)
	go sws.receiveWsMsg(quitChan)
}

// handleSlaveEvent 处理从本地命令读取的事件
func (sws *LocalWsSession) handleSlaveEvent(exitCh chan bool) {
	defer setQuit(exitCh)
	defer log.Printf("thread of handle slave event has exited now")

	buffer := make([]byte, 8192) // 增大缓冲区到 8KB
	for {
		select {
		case <-exitCh:
			return
		default:
			n, _ := sws.slave.Read(buffer)
			_ = sws.masterWrite(buffer[:n])
		}
	}
}

// masterWrite 写入数据到 WebSocket
func (sws *LocalWsSession) masterWrite(data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("A panic occurred during write ws message to master, error message: %v", r)
		}
	}()

	sws.writeMutex.Lock()
	defer sws.writeMutex.Unlock()

	wsData, err := json.Marshal(WsMsg{
		Type: WsMsgCmd,
		Data: base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to encoding to json")
	}

	err = sws.wsConn.WriteMessage(websocket.TextMessage, wsData)
	if err != nil {
		return errors.Wrapf(err, "failed to write to master")
	}
	return nil
}

// receiveWsMsg 接收 WebSocket 消息
func (sws *LocalWsSession) receiveWsMsg(exitCh chan bool) {
	defer func() {
		if r := recover(); r != nil {
			setQuit(exitCh)
			log.Printf("A panic occurred during receive ws message, error message: %v", r)
		}
	}()

	wsConn := sws.wsConn
	defer setQuit(exitCh)
	defer log.Printf("thread of receive ws msg has exited now")

	for {
		select {
		case <-exitCh:
			return
		default:
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("reading websocket message failed, err: %v", err)
				return
			}

			msgObj := WsMsg{}
			_ = json.Unmarshal(wsData, &msgObj)

			switch msgObj.Type {
			case WsMsgResize:
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := sws.slave.ResizeTerminal(msgObj.Cols, msgObj.Rows); err != nil {
						log.Printf("local pty change windows size failed, err: %v", err)
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
					log.Printf("local sending heartbeat to websocket failed, err: %v", err)
				}
			}
		}
	}
}

// sendWebsocketInputCommandToSshSessionStdinPipe 发送命令到本地命令
func (sws *LocalWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.slave.Write(cmdBytes); err != nil {
		log.Printf("ws cmd bytes write to local.stdin pipe failed, err: %v", err)
	}
}