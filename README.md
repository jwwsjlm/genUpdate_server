# 通用更新服务端

用来实现自动更新.客户端版本和服务端存放的版本保持同步.

示例

[网站]: http://up.975135.xyz/updateList/%E6%98%9F%E6%9C%88	"网站"

```json
{"appList":{"fileName":"星月","ReleaseNote":{"appName":"星月","description":"俺只是个测试的软件公告.并无实际功能.不要下载我噢","version":"1.0.0"},"fileList":[{"path":"星月/DllInject.exe","name":"DllInject.exe","size":2686976,"sha256":"e220d39248024bbe54ffc1737b8924711b595cfe4301a72c1483be0522b1b843","downloadURL":"/download/UPIi2w5EjQLGsWHkmsxQd"},{"path":"星月/ReleaseNote.txt","name":"ReleaseNote.txt","size":139,"sha256":"7c0eab3edb699453d2327865945edb6a0ce13a8b2bc61e9768f8fc0679ca0cdd","downloadURL":"/download/vznI2VmWM7MpuVL8ivldx"},{"path":"星月/data/client.dll","name":"client.dll","size":7253584,"sha256":"84986b784d7a263da991d3be04bbafa25e1669453b7b7ad6efdd0abc8547e9af","downloadURL":"/download/QV1xBsdaEOqywTUB4BVHk"},{"path":"星月/data/sql.txt","name":"sql.txt","size":0,"sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","downloadURL":"/download/KlPCceGgXCH_TmLcChwwi"},{"path":"星月/qqwry - 副本.exe","name":"qqwry - 副本.exe","size":25272339,"sha256":"830722bcb86593272040534f993d81bb426096c6adf2e46312e44c31a11745e3","downloadURL":"/download/7hwQB8IZVMDfduCJ4tXBd"},{"path":"星月/qqwry.dat","name":"qqwry.dat","size":25272339,"sha256":"830722bcb86593272040534f993d81bb426096c6adf2e46312e44c31a11745e3","downloadURL":"/download/M_zpjtCu_G8ec39jwohxl"}]},"ret":"ok"}
```

返回update文件夹下的某个文件夹的信息.

## 用法.

1. 部署到自己的服务器当中.

2. update文件夹为软件目录.当中的每个子目录为单独的软件版本

3. 拼接url.比如上传的例子当中.update下的星月软件.

4. 构建自己的公告.在软件目录当中新建ReleaseNote.txt文件.按照例子格式

   

   ```json
   {
     "appName": "星月",
     "description": "This release includes several bug fixes and performance improvements.",
     "version": "1.0.0"
   }
   ```

   存放软件公告.更新版本

   

5. 访问url 

   [示例]: http://up.975135.xyz/updateList/%E6%98%9F%E6%9C%88	"星月软件"

6. downloadURL为每个文件的下载地址.使用 `github.com/matoous/go-nanoid` 包随机生成一串id.使用 `github.com/rosedblabs/rosedb` 储存进db当中.生命周期为10分钟.防止url泄露.被刷流量.拼接url访问实现下载

7. `.ignore` 文件为忽略列表.参考.gitignore语法使用 `github.com/Diogenesoftoronto/go-gitignore` 包实现正则匹配.过滤功能.

8. `jsonbody.json`文件,软件会5分钟同步一次update文件的列表.输出json 文本到这个文件.方便参考

9. 具体的编译方法.放在了`Makefile`文件.参考编译即可

更新客户端.界面写的太丑了.没写.自己写个下载器对比下本地文件的sha256即可.拼接下url即可下载.如果这方面的能人.欢迎提交pr.