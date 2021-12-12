# cloud-native-study

```bash
cd manifests/loki-stack
helm upgrade --install loki ./loki-stack --set grafana.enabled=true,prometheus.enabled=true,prometheus.alertmanager.persistentVolume.enabled=false,prometheus.server.persistentVolume.enabled=false

kubectl apply -f ../httpserver-deployment.yaml

# 修改grafana service type为NodePort
kubectl edit svc loki-grafana

# 查看grafana登录密码
kubectl get secret loki-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo

# 修改prometheus-server service type为NodePort
kubectl edit svc loki-prometheus-server
```

登录prometheus查看指标采集结果：


登录grafana，导入dashboard/httpserver-latency.json并查看：

