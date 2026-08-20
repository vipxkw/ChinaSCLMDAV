# ChinaSCLM DAV · WebDAV 网盘服务器

基于 **Go + Tailwind CSS** 的 WebDAV 网盘服务，内置 iOS 磨砂玻璃风格 Web 界面。
纯 Go 实现（`modernc.org/sqlite`，无 CGO），前端通过 `embed` 打进单个二进制，可打包为
Docker 镜像（支持 `linux/amd64` 与 `linux/arm64`）。

## 功能

- WebDAV 协议（`/dav/`，基于 `golang.org/x/net/webdav`），兼容 RaiDrive / 坚果云 等客户端
- Web 界面：登录 / 仪表盘 / 文件管理 / 分享链接 / 回收站 / 应用密码 / 设置 / 审计日志
- 多用户、会话 Cookie、bcrypt 密码、TOTP 两步验证、应用密码（Basic Auth）
- 删除进回收站、覆盖写保留历史版本、过期分享链接、审计日志
- iOS 风格 UI：白色主背景、磨砂玻璃拟态、大圆角、柔和阴影、半透明导航栏、浅色主题、移动端/桌面端响应式

## 本地运行

```bash
# 1) 编译前端（需要 tailwindcss 独立可执行文件，位于 tool/）
#    源码与编译产物均在 internal/static/，编译后由 Go embed 直接打进二进制
../../tool/tailwindcss -i input.css -o dist/assets/app.css --minify
cp index.html dist/index.html
cp app.js dist/assets/app.js

# 2) 编译 Go 服务（前端已 embed）
go build -o chinasclmdav .

# 3) 运行（首次运行需 --seed-pass 设置管理员密码）
./chinasclmdav -listen :8080 -seed-user xxxx -seed-email xxxx@qq.com -seed-pass '你的密码'
```

打开 http://localhost:8080 登录。

## Docker 构建（多架构 arm64 + amd64）

```bash
# 构建并推送（默认 linux/amd64,linux/arm64）
IMAGE=your-registry/chinasclmdav:v1.0 ./build.sh

# 仅本地构建单架构（不推送）
PUSH=0 PLATFORMS=linux/arm64 IMAGE=chinasclmdav:latest ./build.sh
```

## Docker 部署

直接运行官方镜像：

```bash
docker run -d --name chinasclmdav -p 8080:8080 \
  -v chinasclmdav-data:/data \
  vipiu/chinasclmdav:latest
```

默认种子账号 **admin / admin123**，登录后到「系统设置」修改密码即可（只内置默认值，账号数据保存在 `/data`，不开放未授权登录）。

或使用 Docker Compose：

```bash
docker compose up -d --build
```

默认端口 `8080`，数据（用户文件 + SQLite）保存在 volume `chinasclmdav-data:/data`。
首次启动若需自定义种子账号，可用环境变量覆盖：

| 环境变量 | 说明 | 默认 |
| --- | --- | --- |
| `CHINASCLMDAV_LISTEN` | 监听地址 | `:8080` |
| `CHINASCLMDAV_DATA` | 数据目录 | `/data` |
| `CHINASCLMDAV_PUBLIC_URL` | 对外基础地址（WebDAV 提示用） | `http://localhost:8080` |
| `CHINASCLMDAV_SEED_USER` | 种子账号用户名 | `admin` |
| `CHINASCLMDAV_SEED_EMAIL` | 种子账号邮箱 | `admin@qq.com` |
| `CHINASCLMDAV_SEED_NAME` | 种子账号显示名 | `admin` |
| `CHINASCLMDAV_SEED_PASS` | 种子账号密码（默认内置 `admin123`） | `admin123` |

## 使用 WebDAV 客户端

1. 登录 Web 界面 →「应用密码」→ 添加应用，得到应用密码
2. 客户端服务器地址填 `http://<host>:8080/dav/`
3. 账号用邮箱/用户名，密码用应用密码（不要用登录密码）