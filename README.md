# health-check-exporter
dockerfile 里面的golang和alpine镜像在hubdocker上有，如果无外网的话，需要自行先手动下载

## docker镜像地址
https://hub.docker.com/r/wenyuan1010/health-check-exporter/

## 如何使用？
### 1直接使用docker镜像
只需更改helm-chart里面的values.yaml即可，修改点如下：
* service所在的命名空间（wenyuan-test）
* prometheus所在命名空间（cattle-prometheus）
* 还有你想要监控的healthurls
* 以及镜像地址和版本
* Prometheus定时采集间隔时间（15s）
  ![image](https://github.com/WenYuan1010/health-check-exporter/assets/105798640/da3371cf-f180-4637-ad41-2575c7844f09)

### 2自己使用dockerfile打包镜像
在1修改的基础上再修改一下dockerfile里面的golang和alpine镜像地址（如果有需要的话）

### 3指标解释
|指标| 数据类型 | 备注 |
| --- | --- | --- |
|  application_health| 布尔值（1 正常 0 不正常） | 不同微服务的健康状态可通过url来区分 |
|  system_health| 布尔值（1 正常 0 不正常） | 只要有一个application_health为0，system_health就为0；application_health全为1时system_health=1 |

![image](https://github.com/WenYuan1010/health-check-exporter/assets/105798640/2d6707c0-4b10-403a-863a-d60c809c0b8f)

### 4flag
|flag| 解释 | 备注 |
| --- | --- | --- |
|  -health-urls| 监控微服务地址 |  |
|  -listen-addr| 监听地址 | 默认是:8080 |
|  -timeout-seconds| 超时时间 | 默认是1s,这个是请求healthUrls的超时时间 |

![image](https://github.com/WenYuan1010/health-check-exporter/assets/105798640/383da9dd-651c-4af8-82c9-245a14bc3530)

