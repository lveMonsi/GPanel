//go:build linux

package terminal

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
)

const (
	// DefaultCloseSignal 默认关闭信号
	DefaultCloseSignal = syscall.SIGINT
	// DefaultCloseTimeout 默认关闭超时
	DefaultCloseTimeout = 10 * time.Second
	// closeGracePeriod 优雅关闭等待时间
	closeGracePeriod = 100 * time.Millisecond
)

// LocalCommand 本地命令执行器
type LocalCommand struct {
	closeSignal  syscall.Signal
	closeTimeout time.Duration

	cmd *exec.Cmd
	pty *os.File
}

// NewCommand 创建新的本地命令
func NewCommand(name string, arg ...string) (*LocalCommand, error) {
	cmd := exec.Command(name, arg...)

	if term := os.Getenv("TERM"); term != "" {
		cmd.Env = append(os.Environ(), "TERM="+term)
	} else {
		cmd.Env = append(os.Environ(), "TERM=xterm")
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		cmd.Dir = homeDir
	}

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	lcmd := &LocalCommand{
		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,

		cmd: cmd,
		pty: ptyFile,
	}

	return lcmd, nil
}

// Read 从伪终端读取数据
func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return lcmd.pty.Read(p)
}

// Write 向伪终端写入数据
func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return lcmd.pty.Write(p)
}

// Close 关闭命令和伪终端
func (lcmd *LocalCommand) Close() error {
	if lcmd.pty != nil {
		// 创建带超时的 context
		ctx, cancel := context.WithTimeout(context.Background(), lcmd.closeTimeout)
		defer cancel()

		// 发送 Ctrl+C
		_, _ = lcmd.pty.Write([]byte{3})
		select {
		case <-ctx.Done():
			log.Printf("close timeout after Ctrl+C")
		case <-time.After(closeGracePeriod):
		}

		// 发送 Ctrl+D
		_, _ = lcmd.pty.Write([]byte{4})
		select {
		case <-ctx.Done():
			log.Printf("close timeout after Ctrl+D")
		case <-time.After(closeGracePeriod):
		}

		// 发送 exit 命令
		_, _ = lcmd.pty.Write([]byte("exit\n"))
	}

	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		// 发送 SIGTERM
		_ = lcmd.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-time.After(closeGracePeriod):
			// 如果进程仍在运行，发送 SIGKILL
			if lcmd.cmd.ProcessState == nil || !lcmd.cmd.ProcessState.Exited() {
				_ = lcmd.cmd.Process.Kill()
				log.Printf("process killed after SIGTERM timeout")
			}
		}
	}

	// 关闭伪终端
	if lcmd.pty != nil {
		_ = lcmd.pty.Close()
		lcmd.pty = nil
	}

	return nil
}

// ResizeTerminal 调整终端大小
func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)

	if errno != 0 {
		return errno
	}
	return nil
}

// Wait 等待命令执行完成
func (lcmd *LocalCommand) Wait(quitChan chan bool) {
	if err := lcmd.cmd.Wait(); err != nil {
		log.Printf("local command wait failed, err: %v", err)
	}
	setQuit(quitChan)
}

// setQuit 设置退出信号（安全版本，避免向已关闭的 channel 发送）
func setQuit(ch chan bool) {
	select {
	case ch <- true:
		// 成功发送
	default:
		// channel 已关闭或缓冲区已满，忽略
	}
}
