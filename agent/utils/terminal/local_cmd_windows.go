//go:build windows

package terminal

import (
	"errors"
	"log"
)

const (
	// DefaultCloseSignal 默认关闭信号
	DefaultCloseSignal = 0
	// DefaultCloseTimeout 默认关闭超时
	DefaultCloseTimeout = 0
)

// LocalCommand 本地命令执行器
type LocalCommand struct{}

// NewCommand 创建新的本地命令
func NewCommand(name string, arg ...string) (*LocalCommand, error) {
	return nil, errors.New("local terminal is not supported on Windows, please use SSH terminal instead")
}

// Read 从伪终端读取数据
func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return 0, nil
}

// Write 向伪终端写入数据
func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return 0, nil
}

// Close 关闭命令和伪终端
func (lcmd *LocalCommand) Close() error {
	return nil
}

// ResizeTerminal 调整终端大小
func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	return errors.New("local terminal is not supported on Windows")
}

// Wait 等待命令执行完成
func (lcmd *LocalCommand) Wait(quitChan chan bool) {
	log.Printf("local command wait called on Windows")
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