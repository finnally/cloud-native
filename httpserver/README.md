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
bash run.sh
```

clean.sh 内容：
* 停止并删除容器
* 删除构建的镜像
