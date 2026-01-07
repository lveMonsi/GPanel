package controllers

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func RestartServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Server will restart in 2 seconds",
	})

	// 延迟重启，给响应时间
	go func() {
		time.Sleep(2 * time.Second)
		restart()
	}()
}

func restart() {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// 获取当前参数
	args := os.Args

	// 根据操作系统执行重启
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 cmd /c start 启动新进程
		cmdArgs := []string{"cmd", "/c", "start"}
		cmdArgs = append(cmdArgs, execPath)
		cmdArgs = append(cmdArgs, args[1:]...)
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		// 退出当前进程
		os.Exit(0)
	default:
		// Linux/Mac 使用 exec 替换当前进程
		err = syscall.Exec(execPath, append([]string{execPath}, args[1:]...), os.Environ())
		if err != nil {
			panic(err)
		}
	}
}