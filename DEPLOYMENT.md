# GPanel 部署指南

## 问题说明

当通过 HTTPS 反向代理访问 GPanel 时，浏览器提示"网站的某些部分不安全"，这通常是由于：

1. **缺少安全相关的 HTTP 响应头**
2. **反向代理配置不正确**
3. **前端代码使用了绝对路径的 HTTP 资源**

## 解决方案

### 1. 代码层面修复（已完成）

✅ 已添加安全中间件，自动添加以下 HTTP 响应头：
- `X-Frame-Options: DENY` - 防止点击劫持
- `X-Content-Type-Options: nosniff` - 防止 MIME 类型嗅探
- `X-XSS-Protection: 1; mode=block` - 启用 XSS 过滤器
- `Content-Security-Policy` - 防止混合内容攻击
- `Strict-Transport-Security` - 强制使用 HTTPS
- `Referrer-Policy` - 控制推荐人信息
- `Permissions-Policy` - 控制浏览器权限

✅ 已确保前端 API 请求使用相对路径，自动继承当前协议

### 2. 重新构建应用

```bash
# 在 Linux 服务器上

# 1. 构建前端
cd GPanel/frontend
npm install
npm run build

# 2. 复制构建产物到 core/web/dist
cp -r dist/* ../core/web/dist/

# 3. 构建 Go 应用
cd ../core
go build -o gpanel main.go frontend.go

# 4. 运行应用
./gpanel
```

### 3. Nginx 反向代理配置

参考 `nginx.example.conf` 文件进行配置。关键点：

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL 证书
    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;

    # 重要：传递正确的协议头
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;  # 关键：告诉后端使用 HTTPS
    }
}
```

### 4. 验证修复

1. 重新构建并部署 GPanel
2. 重启 Nginx：`systemctl restart nginx`
3. 访问 HTTPS 站点
4. 打开浏览器开发者工具（F12）
5. 查看 Console 标签页，确认没有混合内容警告
6. 查看 Network 标签页，确认所有资源都通过 HTTPS 加载

### 5. 使用 Let's Encrypt 自动获取证书

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 自动配置 SSL
sudo certbot --nginx -d your-domain.com

# 证书会自动续期
sudo certbot renew --dry-run
```

## 常见问题

### Q: 仍然显示"网站的某些部分不安全"

**A:** 检查以下几点：
1. 确认已重新构建并部署最新代码
2. 检查 Nginx 配置中是否正确设置了 `X-Forwarded-Proto`
3. 清除浏览器缓存并重新加载页面
4. 检查浏览器开发者工具的 Console 是否有混合内容错误

### Q: API 请求失败

**A:** 确认：
1. Nginx 反向代理配置正确
2. GPanel 服务正在运行（端口 8080）
3. 防火墙允许 8080 端口的本地访问

### Q: 如何在生产环境中运行

**A:** 建议使用 systemd 服务：

```ini
# /etc/systemd/system/gpanel.service
[Unit]
Description=GPanel Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/gpanel
ExecStart=/opt/gpanel/gpanel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable gpanel
sudo systemctl start gpanel
sudo systemctl status gpanel
```

## 安全建议

1. **始终使用 HTTPS**：不要在 HTTP 环境中运行生产实例
2. **定期更新**：保持 GPanel 和依赖库的最新版本
3. **使用强密码**：如果添加了认证功能，使用强密码
4. **配置防火墙**：只开放必要的端口（80, 443）
5. **定期备份**：备份配置文件和数据
6. **监控日志**：定期检查应用和 Nginx 日志

## 技术支持

如遇到问题，请检查：
1. 应用日志：`/opt/1panel/log/gpanel.log`（如果配置了日志）
2. Nginx 日志：`/var/log/nginx/gpanel_error.log`
3. 浏览器开发者工具的 Console 和 Network 标签页