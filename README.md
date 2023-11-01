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
|  system_health| 布尔值（1 正常 0 不正常） | 针对url带有health的才会根据标签来输出系统的健康状态，系统里面所有的服务中只要有一个不是1，system_health就是0|

例如针对标签iot2,共有两个微服务，http://iot-auth-cs.iot2/actuator/health,http://iot-auth.iot2/actuator/health,只要有一个不健康，system_health{label="iot2"}就是0

![image](https://github.com/WenYuan1010/health-check-exporter/assets/105798640/f84add63-95f5-44fe-93ef-03d7c6a830f2)
![image](https://github.com/WenYuan1010/health-check-exporter/assets/105798640/5554e830-e2b2-4d99-9fb6-f5ffcd5d33a2)

### 4http探针
有的系统并没有健康检查的配置，没办法按照/actuator/health的方式去采集健康状态，只能退而求其次地使用http探针去探测，如果返回200，则认为是健康，反之不健康。只要url不带/actuator/health都会通过http探针去输出服务的健康状态
### 5flag
|flag| 解释 | 备注 |
| --- | --- | --- |
|  -health-urls| 监控微服务地址 |  |
|  -listen-addr| 监听地址 | 默认是:8080 |
|  -timeout-seconds| 超时时间 |单位为s; 默认是1s,这个是请求healthUrls的超时时间 |
|  -labels| 系统标签 | 标签指的是集群里面的命名空间 |

go run .\main.go -labels="iot2:,ioms-alm:" -health-urls="http://iot-auth-cs.iot2/actuator/health,http://iot-auth.iot2/actuator/health,http://ioms-alarm-ioms3-alm-auth-cs.ioms-alm.svc.cluster.local:8080/actuator/health,http://ioms-alarm-alm-auth-cs.ioms-alm.svc.cluster.local:8080/actuator/health,http://10.4.1.156:30159,http://10.4.1.156:30158" -timeout-seconds=1 -listen-addr=":9999"

