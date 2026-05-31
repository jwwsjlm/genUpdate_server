# 通用更新服务端

一个轻量的自动更新服务端，用于集中管理多个软件的版本清单和更新文件。服务会扫描更新目录，生成带 SHA256 校验值的文件清单，并提供清单查询、文件下载、Web 查看、健康检查和构建版本接口。

[![GitHub Release](https://img.shields.io/github/v/release/jwwsjlm/genUpdate_server)](https://github.com/jwwsjlm/genUpdate_server/releases)
[![Go Version](https://img.shields.io/badge/go-1.26.3-blue)](https://golang.org)

## 功能

- 多软件管理：`update` 目录下的每个一级子目录对应一个软件。
- 自动生成清单：启动时扫描更新目录，并按配置的间隔定时刷新。
- SHA256 校验：为每个更新文件生成 SHA256，客户端可用于完整性校验。
- 扫描缓存：通过 `manifest-cache.json` 缓存文件大小、修改时间和 SHA256，未变化文件不会重复计算哈希。
- Web 更新中心：访问 `/` 可查看软件、版本说明、文件列表、SHA256 和下载入口。
- Token 生成器：Web 页面可本地生成随机 token，方便写入 `appTokens`。
- 私有分发：支持按软件配置 token，避免一个用户下载其他软件的文件。
- Web 访问保护：支持 bcrypt 密码哈希保护 Web 页面和 `/api/apps`。
- 登录限速：Web 登录接口使用 `golang.org/x/time/rate` 做单 IP 限速，降低密码爆破风险。
- 更新清单签名：可用 Ed25519 私钥为 `/updateList/:app` 响应签名，方便客户端验证清单未被篡改。
- 下载白名单：下载接口只允许访问已经进入更新清单的文件。
- 断点续传：下载接口支持 `Range`、`HEAD`、`Accept-Ranges` 和 `ETag`。
- 并发限制：支持全局下载并发和单 IP 下载并发限制。
- Docker 部署：内置多阶段 Dockerfile，可直接挂载更新目录运行。

## 快速开始

源码运行：

```bash
git clone https://github.com/jwwsjlm/genUpdate_server.git
cd genUpdate_server
go run ./cmd/main
```

默认访问地址：

```text
http://localhost:8090
```

浏览器打开 `http://localhost:8090/` 可进入 Web 更新中心。

## 编译

```bash
go build -o genupdate-server ./cmd/main
./genupdate-server
```

也可以使用 Makefile：

```bash
make build-windows
make build-linux
```

## Docker 部署

```bash
docker run -d \
  --name genupdate-server \
  -p 8090:8090 \
  -v ./update:/app/update \
  -v ./log:/app/log \
  jwwsjlm/genUpdate_server:latest
```

容器内默认更新目录为 `/app/update`，默认监听 `8090`。

## 更新目录结构

```text
update/
├── .ignore
├── cc/
│   ├── ReleaseNote.txt
│   ├── app.exe
│   └── data/
│       └── sqlite.sqlite
└── bb/
    ├── ReleaseNote.txt
    └── app.exe
```

每个一级目录会作为一个软件名称。例如 `cc` 对应：

```text
/updateList/cc
/download/cc/app.exe
```

### ReleaseNote.txt

每个软件目录可放置 `ReleaseNote.txt`，内容为 JSON：

```json
{
  "appName": "软件名称",
  "description": "更新说明",
  "version": "1.0.0"
}
```

如果没有提供，服务会使用默认值：

```json
{
  "appName": "目录名",
  "description": "null",
  "version": "1.0.0"
}
```

### .ignore

在更新根目录创建 `update/.ignore`，可按 gitignore 风格忽略不需要进入清单的文件。

服务还会自动忽略以下内部文件：

- `ReleaseNote.txt`
- `.ignore`
- `jsonBody.json`
- `manifest-cache.json`
- 隐藏文件和隐藏目录，例如 `.env`、`.secret/`

## 配置

服务启动时会先读取默认配置，再读取本地配置文件，最后应用环境变量覆盖。

优先级：

```text
默认值 < config.json < 环境变量
```

默认会尝试读取当前工作目录下的 `config.json`。也可以通过 `GENUPDATE_CONFIG` 指定配置文件路径。

### config.json 示例

```json
{
  "port": "8090",
  "updateDir": "update",
  "scanIntervalSeconds": 300,
  "readTimeoutSeconds": 15,
  "writeTimeoutSeconds": 600,
  "idleTimeoutSeconds": 60,
  "maxConcurrentDownloads": 64,
  "maxConcurrentDownloadsPerIP": 8,
  "webPasswordHash": "$2a$10$replace-with-bcrypt-hash",
  "webSessionSecret": "replace-with-random-session-secret",
  "manifestSigningPrivateKey": "replace-with-generated-ed25519-private-seed",
  "manifestSigningKeyID": "replace-with-signing-key-id",
  "appTokens": {
    "cc": "cc-token",
    "bb": "bb-token"
  }
}
```

`updateDir` 使用相对路径时，会相对于服务工作目录解析。

### 环境变量

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `GENUPDATE_CONFIG` | 当前工作目录下的 `config.json` | 本地配置文件路径 |
| `GENUPDATE_PORT` | `8090` | HTTP 监听端口，支持 `8090` 或 `:8090` |
| `GENUPDATE_UPDATE_DIR` | 当前工作目录下的 `update` | 更新文件根目录 |
| `GENUPDATE_SCAN_INTERVAL_SECONDS` | `300` | 更新目录扫描间隔，单位秒 |
| `GENUPDATE_READ_TIMEOUT_SECONDS` | `15` | HTTP 读取超时，单位秒 |
| `GENUPDATE_WRITE_TIMEOUT_SECONDS` | `600` | HTTP 写入超时，单位秒 |
| `GENUPDATE_IDLE_TIMEOUT_SECONDS` | `60` | HTTP 空闲连接超时，单位秒 |
| `GENUPDATE_MAX_CONCURRENT_DOWNLOADS` | `64` | 最大并发下载数 |
| `GENUPDATE_MAX_CONCURRENT_DOWNLOADS_PER_IP` | `8` | 单个客户端 IP 最大并发下载数 |
| `GENUPDATE_APP_TOKENS` | 空 | 按软件授权的 token 映射，例如 `cc=cc-token,bb=bb-token` |
| `GENUPDATE_WEB_PASSWORD_HASH` | 空 | Web 管理页面 bcrypt 密码哈希；配置后访问 Web 和 `/api/apps` 需要登录 |
| `GENUPDATE_WEB_SESSION_SECRET` | `GENUPDATE_WEB_PASSWORD_HASH` | Web 登录 cookie 签名密钥，建议使用随机长字符串 |
| `GENUPDATE_MANIFEST_SIGNING_PRIVATE_KEY` | 空 | Ed25519 私钥种子或私钥，支持 base64url、base64、hex；配置后更新清单会带签名 |
| `GENUPDATE_MANIFEST_SIGNING_KEY_ID` | 自动生成 | 清单签名 key id，方便客户端识别当前公钥 |

## Web 密码

生成 Web 管理密码哈希：

```bash
genupdate-server hash-password "your-admin-password"
```

将输出写入 `GENUPDATE_WEB_PASSWORD_HASH` 或 `config.json` 的 `webPasswordHash`。

`webSessionSecret` 用于签名登录 cookie，不是登录密码，也不是 bcrypt salt。建议单独配置一个随机长字符串。更换后，旧的网页登录状态会失效，需要重新登录。

## 更新清单签名

生成 Ed25519 签名密钥：

```bash
genupdate-server generate-signing-key
```

输出示例：

```text
GENUPDATE_MANIFEST_SIGNING_PRIVATE_KEY=...
GENUPDATE_MANIFEST_SIGNING_PUBLIC_KEY=...
GENUPDATE_MANIFEST_SIGNING_KEY_ID=...
```

服务端只需要保存 `GENUPDATE_MANIFEST_SIGNING_PRIVATE_KEY` 和 `GENUPDATE_MANIFEST_SIGNING_KEY_ID`。客户端应内置或配置 `GENUPDATE_MANIFEST_SIGNING_PUBLIC_KEY`，用于验证 `/updateList/:app` 返回的签名。

配置私钥后，更新清单会额外返回：

```json
{
  "signature": "...",
  "signatureAlgorithm": "ed25519",
  "signatureKeyID": "..."
}
```

未配置私钥时，接口保持旧响应格式，不影响现有客户端。

## API

### 健康检查

```bash
curl http://localhost:8090/healthz
```

返回示例：

```json
{
  "ret": "ok",
  "status": "healthy"
}
```

### 查看服务版本

```bash
curl http://localhost:8090/version
```

返回内容包含构建版本、提交号、构建时间、当前清单大小和清单缓存时间。

### 获取软件更新清单

```bash
curl http://localhost:8090/updateList/cc
```

返回示例：

```json
{
  "ret": "ok",
  "appList": {
    "fileName": "cc",
    "ReleaseNote": {
      "appName": "cc",
      "description": "更新说明",
      "version": "1.0.0"
    },
    "fileList": [
      {
        "path": "cc/app.exe",
        "name": "app.exe",
        "size": 12345,
        "sha256": "...",
        "downloadURL": "/download/cc/app.exe",
        "modTime": "2026-05-30T00:00:00Z"
      }
    ]
  }
}
```

### 下载文件

```bash
curl -L "http://localhost:8090/download/cc/app.exe" -o app.exe
```

### 获取 Web 软件列表

```bash
curl http://localhost:8090/api/apps
```

该接口供 Web 页面使用。配置 `GENUPDATE_WEB_PASSWORD_HASH` 后需要先登录。配置 `GENUPDATE_APP_TOKENS` 后，未登录时只会返回 token 对应的软件。

## 私有分发

如果服务端挂了多个软件，但每个用户只能下载自己对应的软件，建议配置 `appTokens`。

示例：

```bash
GENUPDATE_APP_TOKENS="cc=cc-secret,bb=bb-secret" ./genupdate-server
```

访问 `cc`：

```bash
curl -H "Authorization: Bearer cc-secret" http://localhost:8090/updateList/cc
curl -H "Authorization: Bearer cc-secret" http://localhost:8090/download/cc/app.exe -o app.exe
```

此时即使用户知道 `bb` 的名称，用 `cc-secret` 访问 `/updateList/bb` 或 `/download/bb/...` 也只会得到 404，避免暴露其他软件信息。

客户端也可以使用：

```text
X-Update-Token: cc-secret
```

## 安全说明

- 下载接口会校验路径，阻止 `../` 等路径穿越。
- 下载接口只允许访问当前清单中的文件，内部文件即使存在也不能通过 `/download/*` 下载。
- 扫描时默认跳过隐藏文件和隐藏目录，避免 `.env`、`.secret/` 等敏感文件误进入更新清单。
- 配置 `GENUPDATE_APP_TOKENS` 后，清单和下载接口会按软件 token 授权；错误 token 返回 404，减少软件名称泄露。
- 配置 `GENUPDATE_WEB_PASSWORD_HASH` 后，Web 页面和 `/api/apps` 需要登录。
- Web 密码使用 bcrypt 哈希保存，bcrypt 自带 salt，不需要保存明文密码。
- Web 登录 cookie 使用 HMAC-SHA256 签名，`webSessionSecret` 应使用随机长字符串。
- Web 登录接口默认限制单 IP 每分钟 5 次尝试，超过后返回 429。
- 建议启用更新清单签名，客户端验证签名后再下载和替换文件，降低清单被篡改的风险。
- 服务使用 `golang.org/x/sync/semaphore` 限制全局并发下载数和单 IP 并发下载数，避免单个客户端多线程下载占满服务端连接。
- 对外部署时建议只把可公开分发的更新包放进 `update` 目录，并在反向代理层启用 HTTPS、访问日志、限速和必要的鉴权策略。

## 性能说明

- 文件 SHA256 只在文件大小或修改时间变化时重新计算。
- 下载文件校验使用清单路径索引，避免每次下载遍历全部文件。
- 文件下载由 Go 标准库 `http.ServeContent` 处理，支持 Range 和 HEAD。
- 大文件下载会占用连接，建议根据机器带宽和磁盘能力调整 `maxConcurrentDownloads` 和 `maxConcurrentDownloadsPerIP`。

## 发布

项目使用 GoReleaser 自动构建发布包。创建并推送 `vMAJOR.MINOR.PATCH` 格式的 tag 后，GitHub Actions 会自动测试、构建 Linux/Windows 的 amd64/arm64 产物，并生成 checksums。

```bash
git tag -a v0.2.1 -m "v0.2.1"
git push origin v0.2.1
```

也可以在 GitHub Actions 的 Release workflow 里手动输入 tag 触发发布。

## 相关项目

- 客户端：[genUpdate_client](https://github.com/jwwsjlm/genUpdate_client)

## 许可证

MIT License
