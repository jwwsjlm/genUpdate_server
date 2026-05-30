# 通用更新服务端

一个轻量的自动更新服务端，用于集中管理多个软件的版本清单和更新文件。服务会扫描更新目录，生成带 SHA256 校验值的文件清单，并提供清单查询、文件下载、健康检查和构建版本信息接口。

[![GitHub Release](https://img.shields.io/github/v/release/jwwsjlm/genUpdate_server)](https://github.com/jwwsjlm/genUpdate_server/releases)
[![License](https://img.shields.io/github/license/jwwsjlm/genUpdate_server)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26.3-blue)](https://golang.org)

---

## 功能介绍

- 多软件管理：`update` 目录下每个一级子目录对应一个软件，单个服务即可维护多个软件的更新清单。
- 自动清单生成：启动时扫描更新目录，并按配置的间隔定时刷新清单。
- SHA256 校验：为每个更新文件生成 SHA256，客户端可用它校验下载文件完整性。
- 扫描缓存：通过 `manifest-cache.json` 缓存文件大小、修改时间和 SHA256，未变化文件不会重复计算哈希。
- 稳定下载地址：清单中的 `downloadURL` 使用文件相对路径，例如 `/download/星月/qqwry.dat`。
- 断点续传支持：下载接口支持 `Range` 请求、`HEAD` 请求、`Accept-Ranges` 和 `ETag`。
- 并发下载限制：可通过环境变量限制同时下载数，避免单机资源被打满。
- 忽略规则：支持 `update/.ignore`，并自动跳过 `ReleaseNote.txt`、`.ignore`、`jsonBody.json`、`manifest-cache.json` 等内部文件。
- 版本公告：每个软件目录可放置 `ReleaseNote.txt`，用于返回应用名、版本号和更新说明。
- 运维接口：提供 `/healthz` 健康检查和 `/version` 构建版本信息接口。
- Web 更新中心：访问根路径即可浏览软件、版本公告、文件列表、SHA256 和下载入口。
- 下载白名单：下载接口只允许访问已进入更新清单的文件，内部文件和隐藏文件不会被直接下载。
- Docker 部署：内置多阶段 Dockerfile，可直接挂载更新目录运行。

---

## 快速开始

### 源码运行

```bash
git clone https://github.com/jwwsjlm/genUpdate_server.git
cd genUpdate_server
go run ./cmd/main
```

默认访问地址：`http://localhost:8090`

浏览器打开 `http://localhost:8090/` 可以进入 Web 更新中心。

页面右上角可以直接生成随机 Token，适合复制到 `GENUPDATE_APP_TOKENS` 中使用。

### 编译

```bash
go build -o genUpdate_server ./cmd/main
./genUpdate_server
```

也可以使用 Makefile 交叉编译：

```bash
make build-windows
make build-linux
```

### Docker 部署

```bash
docker run -d \
  -p 8090:8090 \
  -v ./update:/app/update \
  jwwsjlm/genUpdate_server:latest
```

---

## 发布

项目使用 GoReleaser 自动构建发布包。创建并推送符合 `vMAJOR.MINOR.PATCH` 格式的 tag 后，GitHub Actions 会自动测试、构建 Linux/Windows 的 amd64/arm64 产物，并生成 checksums。

```bash
git tag -a v0.2.1 -m "v0.2.1"
git push origin v0.2.1
```

也可以在 GitHub Actions 的 Release workflow 里手动输入 tag 触发发布。

---

## API 使用

### 获取全部软件清单

```bash
curl http://localhost:8090/api/apps
```

该接口供 Web 页面使用，会返回所有软件清单以及软件数量、文件数量和文件总大小。

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
curl http://localhost:8090/updateList/星月
```

返回示例：

```json
{
  "ret": "ok",
  "appList": {
    "fileName": "星月",
    "ReleaseNote": {
      "appName": "星月",
      "description": "更新说明",
      "version": "1.0.0"
    },
    "fileList": [
      {
        "path": "星月/qqwry.dat",
        "name": "qqwry.dat",
        "size": 12345,
        "sha256": "...",
        "downloadURL": "/download/星月/qqwry.dat",
        "modTime": "2026-05-30T00:00:00Z"
      }
    ]
  }
}
```

### 下载文件

```bash
curl -L "http://localhost:8090/download/星月/qqwry.dat" -o qqwry.dat
```

下载接口会限制路径穿越，并只允许访问更新目录内的文件。

---

## 配置

服务启动时会先读取默认配置，再读取本地配置文件，最后应用环境变量覆盖。优先级为：默认值 < `config.json` < 环境变量。

默认会尝试读取当前工作目录下的 `config.json`；也可以通过 `GENUPDATE_CONFIG` 指定配置文件路径。仓库内提供了 `config.example.json`，复制后改名为 `config.json` 即可使用。

### config.json

```json
{
  "port": "8090",
  "updateDir": "update",
  "scanIntervalSeconds": 300,
  "readTimeoutSeconds": 15,
  "writeTimeoutSeconds": 600,
  "idleTimeoutSeconds": 60,
  "maxConcurrentDownloads": 64,
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
| `GENUPDATE_APP_TOKENS` | 空 | 按软件授权的 token 映射，例如 `cc=cc-token,bb=bb-token` |

---

## 更新目录结构

```text
update/
├── .ignore
├── 星月/
│   ├── ReleaseNote.txt
│   ├── qqwry.dat
│   └── data/
│       └── sqlite.sqlite
└── 鬼泣/
    ├── ReleaseNote.txt
    └── demo.exe
```

每个一级目录会作为一个软件名称，例如 `星月` 对应接口 `/updateList/星月`。该目录下的普通文件会进入清单，子目录文件也会保留相对路径。

### ReleaseNote.txt

在软件目录下创建 `ReleaseNote.txt`，内容为 JSON：

```json
{
  "appName": "软件名称",
  "description": "更新说明",
  "version": "1.0.0"
}
```

如果没有提供该文件，服务会使用默认值：

```json
{
  "appName": "目录名",
  "description": "null",
  "version": "1.0.0"
}
```

### .ignore

在更新根目录下创建 `.ignore`，可按 gitignore 风格忽略不需要进入清单的文件。

服务还会自动忽略以下内部文件：

- `ReleaseNote.txt`
- `.ignore`
- `jsonBody.json`
- `manifest-cache.json`

---

## 生成文件

服务扫描后会在更新目录下生成：

- `jsonBody.json`：当前所有软件的更新清单快照。
- `manifest-cache.json`：SHA256 缓存，用于减少重复哈希计算。

这两个文件属于服务内部文件，默认不会出现在更新清单中。

---

## 安全说明

- 下载接口会校验路径，阻止 `../` 等路径穿越。
- 下载接口只服务当前清单中的文件，`jsonBody.json`、`manifest-cache.json`、`.ignore`、`ReleaseNote.txt` 等内部文件即使存在也不能通过 `/download/*` 下载。
- 扫描时默认跳过隐藏文件和隐藏目录，例如 `.env`、`.secret/`，避免敏感文件误进入更新清单。
- 如需私有分发，可配置 `GENUPDATE_APP_TOKENS`。配置后，客户端必须通过 `Authorization: Bearer <token>` 或 `X-Update-Token: <token>` 访问对应软件；错误 token 会返回 404，避免暴露其他软件名称。
- 对外部署时建议只把可公开分发的更新包放进 `update` 目录，并在反向代理层启用 HTTPS、访问日志、限速和必要的鉴权策略。

按软件授权示例：

```bash
GENUPDATE_APP_TOKENS="cc=cc-secret,bb=bb-secret" ./genUpdate_server
curl -H "Authorization: Bearer cc-secret" http://localhost:8090/updateList/cc
curl -H "Authorization: Bearer cc-secret" http://localhost:8090/download/cc/app.exe -o app.exe
```

此时即使用户猜到 `bb`，使用 `cc-secret` 访问 `/updateList/bb` 或 `/download/bb/...` 也只会得到 404。

Web 更新中心提供本地随机 Token 生成按钮，生成逻辑在浏览器内完成，服务端不会保存或记录该 Token。

---

## 相关项目

- 客户端：[genUpdate_client](https://github.com/jwwsjlm/genUpdate_client)
- 依赖：[go-gitignore](https://github.com/matoous/go-gitignore)

---

## 许可证

MIT License
