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

* 以下部分为手动测试：
// 生成一个10M文件，并将文件内容追加到httpserver日志中
root@k8snode:~# dd if=/dev/zero of=/tmp/testfile count=10240 bs=1024
10240+0 records in
10240+0 records out
10485760 bytes (10 MB, 10 MiB) copied, 0.0249115 s, 421 MB/s

root@k8snode:~# cat /tmp/testfile >> $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))

root@k8snode:~# ls -lh $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))
-rw-r----- 1 root root 11M Nov 28 10:18 /var/lib/docker/containers/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6-json.log

// 因为脚本生成的配置是日志文件大小超过10M后进行切分，执行logrotate命令，检查切分是否成功。json.log.1后缀的文件即为切分后保存的日志
root@k8snode:~# logrotate -f /etc/logrotate.d/httpserver

root@k8snode:~# ls -lh $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))
-rw-r----- 1 root root 0 Nov 28 10:19 /var/lib/docker/containers/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6-json.log

root@k8snode:~# ls -lh $(docker inspect -f {{.LogPath}} $(docker ps | awk '/httpserver/&&!/pause/{print $1}' | head -1))*
-rw-r----- 1 root root 120 Nov 28 10:24 /var/lib/docker/containers/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6-json.log
-rw-r----- 1 root root 11M Nov 28 10:24 /var/lib/docker/containers/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6/8abc6a21d7d93b56f7496f2e486070d246a108d1c638636e527a05957a7badf6-json.log.1

实际生产环境中需要根据业务情况调整logrotate配置，可选择按天或者按文件大小两种方式进行日志切分，同时还可以设置是否压缩，保存个数等配置。
由于logrotate默认按天执行，当执行条件不能满足需求时，可以通过设置crontab以期望的时间间隔来执行。
***
```
bash deploy.sh clean
```
功能：
* 删除所有deploy创建的对象