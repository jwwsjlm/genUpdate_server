# 通用更新服务端

📦 自动更新服务端，支持多软件版本管理和 SHA256 校验。

[![GitHub Release](https://img.shields.io/github/v/release/jwwsjlm/genUpdate_server)](https://github.com/jwwsjlm/genUpdate_server/releases)
[![License](https://img.shields.io/github/license/jwwsjlm/genUpdate_server)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org)

---

## ✨ 功能特性

- 🔄 **自动版本同步** - 客户端自动检测并更新
- 📂 **多软件支持** - 单服务端管理多软件
- 🔐 **SHA256 校验** - 确保文件完整性
- ⚡ **稳定下载链接** - 直接按相对路径下载，不依赖临时随机 ID
- 📊 **实时公告** - 支持版本更新说明
- 🚀 **扫描缓存** - 未变化文件不会重复计算 SHA256

---

## 🚀 快速开始

### 源码编译

```bash
git clone https://github.com/jwwsjlm/genUpdate_server.git
cd genUpdate_server
go build -o genUpdate_server ./cmd/main
./genUpdate_server
```

默认访问地址：`http://localhost:8090`

### Docker 部署

```bash
docker run -d -p 8090:8090 -v ./update:/app/update jwwsjlm/genUpdate_server:latest
```

---

## 📖 API 使用

### 获取软件版本

```bash
curl http://localhost:8090/updateList/软件名
```

### 下载文件

```bash
curl -L "http://localhost:8090/download/星月/qqwry.dat" -o qqwry.dat
```

---

## ⚙️ 配置

### 环境变量

- `GENUPDATE_PORT`：监听端口，默认 `8090`
- `GENUPDATE_SCAN_INTERVAL_SECONDS`：扫描间隔秒数，默认 `300`

### ReleaseNote.txt

在软件目录下创建版本公告：

```json
{
  "appName": "软件名称",
  "description": "更新说明",
  "version": "1.0.0"
}
```

---

## 📂 目录结构示例

```text
update/
├── .ignore
├── 星月/
│   ├── ReleaseNote.txt
│   └── qqwry.dat
└── 鬼泣/
    ├── ReleaseNote.txt
    └── demo.exe
```

---

## 🙏 致谢

- [go-gitignore](https://github.com/matoous/go-gitignore)

相关项目：[genUpdate_client](https://github.com/jwwsjlm/genUpdate_client)

---

## 📄 许可证

MIT License

---

**如果有帮助，欢迎 Star ⭐️！**
