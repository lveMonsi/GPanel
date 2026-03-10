package firewall

import (
	"fmt"
	"os/exec"
	"strings"
)

// checkCommand 检查命令是否存在
func checkCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Exec 执行命令
func Exec(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command '%s %v' failed: %w, output: %s", name, args, err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// ExecWithSudo 使用sudo执行命令
func ExecWithSudo(args ...string) (string, error) {
	cmdArgs := append([]string{"-n"}, args...)
	output, err := Exec("sudo", cmdArgs...)
	if err != nil {
		return "", fmt.Errorf("sudo command failed: %w", err)
	}
	return output, nil
}

// RunCommandWithOutput 执行命令并返回输出
func RunCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))
	if err != nil {
		return result, fmt.Errorf("command '%s %v' failed: %w", name, args, err)
	}
	return result, nil
}