# http-server
```
bash run.sh
```

run.sh 内容：
* 构建httpserver镜像
* 推送该镜像到dockerhub
* 运行该镜像，以 -P 方式暴露端口
* 通过inspect找到容器pid，用nsenter命令查看容器内IP配置

```
bash clean.sh
```

clean.sh 内容：
* 停止并删除容器
* 删除构建的镜像


# 部署部分
## part 1
```
bash deploy.sh deploy
```
功能：
* 部署 httpserver-deployment.yaml
* 部署 ingress-deployment.yaml
* 生成SSL证书并创建secret (secret.yaml)
* 创建ingress (ingress.yaml)

**deploy过程中可能会因为ingress-controller未完全启动导致ingress.yaml创建失败，此时需要手动执行 kubectl apply -f ingress.yaml 来创建。**
***
## part 2
```
bash deploy.sh accessTest
```
功能：
* 通过 ingress service clusterIp 访问httpserver服务
* 通过主机名访问httpserver服务
***
```
bash deploy.sh logrotateConfig
```
功能：
* 配置httpserver的logrotate实现日志切分

以下部分为手动测试：

1. 生成一个10M文件，并将文件内容追加到httpserver日志中
  
```
dd if=/dev/zero of=/tmp/testfile count=10240 bs=1024 
cat /tmp/testfile >> $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))  
ls -lh $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))
```

2. 因为脚本生成的配置是日志文件大小超过10M后进行切分，执行logrotate命令，检查切分是否成功。json.log.1后缀的文件即为切分后保存的日志  
  
```
logrotate -f /etc/logrotate.d/httpserver
ls -lh $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))* 
```
  
**实际生产环境中需要根据业务情况调整logrotate配置，可选择按天或者按文件大小两种方式进行日志切分，同时还可以设置是否压缩，保存个数等配置。
由于logrotate默认按天执行，当执行条件不能满足需求时，可以通过设置crontab以期望的时间间隔来执行。**
***
## part 3
```
bash deploy.sh clean
```
功能：
* 删除所有deploy创建的对象