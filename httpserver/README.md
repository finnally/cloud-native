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

```
bash deploy.sh deploy
```
* 部署 httpserver-deployment.yaml
* 部署 ingress-deployment.yaml
* 生成SSL证书并创建secret (secret.yaml)
* 创建ingress (ingress.yaml)

```
bash deploy.sh accessTest
```
* 通过 ingress service clusterIp 访问httpserver服务
* 通过主机名访问httpserver服务

```
bash deploy.sh clean
```
* 删除所有deploy创建的对象