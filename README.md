# health-check-exporter
dockerfile 里面的golang和alpine镜像在hubdocker上有，如果无外网的话，需要自行先手动下载
**这个exporter是配合rancher使用的**

## docker镜像地址
https://hub.docker.com/r/wenyuan1010/health-check-exporter/

## 如何使用？
### 1直接使用docker镜像
只需更改helm-chart里面的几个文件即可，分别是service所在的命名空间（wenyuan-test）和prometheus所在命名空间（cattle-prometheus），还有你想要监控的healthurls.
### 2自己使用dockerfile打包镜像
在1修改的基础上再修改一下dockerfile里面的golang和alpine镜像地址（如果有需要的话）
