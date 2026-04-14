package middleware

import (
	"fmt"
	"gpanel/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 路由到资源/操作的映射
var routeMapping = map[string]struct{ resource, action string }{
	"POST /api/v1/auth/login":          {"auth", "login"},
	"POST /api/v1/hosts":               {"host", "create"},
	"PUT /api/v1/hosts":                {"host", "update"},
	"DELETE /api/v1/hosts":             {"host", "delete"},
	"POST /api/v1/hosts/move":          {"host", "move"},
	"POST /api/v1/hosts/import":        {"host", "import"},
	"POST /api/v1/host-groups":         {"host_group", "create"},
	"PUT /api/v1/host-groups":          {"host_group", "update"},
	"DELETE /api/v1/host-groups":       {"host_group", "delete"},
	"POST /api/v1/cronjobs":            {"cronjob", "create"},
	"POST /api/v1/cronjobs/update":     {"cronjob", "update"},
	"POST /api/v1/cronjobs/delete":     {"cronjob", "delete"},
	"POST /api/v1/cronjobs/toggle":     {"cronjob", "toggle"},
	"POST /api/v1/cronjobs/handle":     {"cronjob", "execute"},
	"POST /api/v1/settings":            {"setting", "create"},
	"PUT /api/v1/settings":             {"setting", "update"},
	"DELETE /api/v1/settings":          {"setting", "delete"},
	"POST /api/v1/settings/system":     {"setting", "update"},
	"POST /api/v1/settings/terminal":   {"setting", "update"},
	"POST /api/v1/agent/firewall/operate": {"firewall", "operate"},
	"POST /api/v1/agent/firewall/port":    {"firewall", "port_rule"},
	"POST /api/v1/agent/firewall/ip":      {"firewall", "ip_rule"},
	"POST /api/v1/agent/firewall/forward": {"firewall", "forward_rule"},
	"POST /api/v1/agent/process/stop":     {"process", "stop"},
	"POST /api/v1/agent/ssh/operate":      {"ssh", "operate"},
	"POST /api/v1/agent/ssh/config":       {"ssh", "config"},
	"POST /api/v1/agent/ssh/keys":         {"ssh", "create_key"},
	"POST /api/v1/agent/ssh/keys/delete":  {"ssh", "delete_key"},
	"POST /api/v1/quick-commands":          {"quick_command", "create"},
	"POST /api/v1/quick-commands/update":   {"quick_command", "update"},
	"POST /api/v1/quick-commands/delete":   {"quick_command", "delete"},
	"POST /api/v1/logs/operation/clean":    {"log", "clean"},
	"POST /api/v1/server/restart":          {"system", "restart"},
	"POST /api/v1/config":                  {"system", "config"},
}

// OperationLog 操作日志记录中间件
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行请求
		c.Next()

		// 仅记录写操作
		method := c.Request.Method
		if method == http.MethodGet {
			return
		}

		// 匹配路由
		path := c.FullPath()
		if path == "" {
			return
		}

		// 去掉路径参数部分进行匹配 (如 /api/v1/hosts/:id -> /api/v1/hosts)
		key := method + " " + path
		mapping, ok := routeMapping[key]
		if !ok {
			// 尝试去掉最后一段路径参数
			if idx := strings.LastIndex(path, "/:"); idx != -1 {
				key = method + " " + path[:idx]
				mapping, ok = routeMapping[key]
			}
			if !ok {
				return
			}
		}

		username, _ := c.Get("username")
		usernameStr := fmt.Sprintf("%v", username)
		if usernameStr == "<nil>" {
			usernameStr = ""
		}

		ip := c.ClientIP()
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failed"
		}

		detail := fmt.Sprintf("%s %s", method, c.Request.URL.Path)

		// 异步发送到 Agent
		go func() {
			client, err := utils.NewAgentClient()
			if err != nil {
				log.Printf("operation log: failed to create agent client: %v", err)
				return
			}
			body := map[string]string{
				"username": usernameStr,
				"ip":       ip,
				"resource": mapping.resource,
				"action":   mapping.action,
				"detail":   detail,
				"status":   status,
			}
			_, _, err = client.RequestWithStatus(http.MethodPost, "/api/v1/logs/operation/create", body)
			if err != nil {
				log.Printf("operation log: failed to record: %v", err)
			}
		}()
	}
}
