# istio发布httpserver服务

## 一、httpserver 服务以 Istio Ingress Gateway 的形式发布
* 安装istio
```
curl -L https://istio.io/downloadIstio | sh -<br>
cd istio-${ISTIO_VERSION}<br>
cp bin/istioctl /usr/local/bin<br>
istioctl install --set profile=demo -y<br>
INGRESS_IP=$(kubectl get svc -nistio-system | awk '/istio-ingressgateway/{print $3}')
```
* 创建namespace: secns<br>
kubectl create ns secns<br>
* 给该ns打上istio-injection标签，使创建pod时自动注入istio-proxy sidecar<br>
kubectl label ns secns istio-injection=enabled<br>
* 在namespace: secns下创建pod和service<br>
kubectl create -f httpserver.yaml -n secns<br>
* 生成SSL证书<br>
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=cns Inc./CN=*.cns.io' -keyout cns.io.key -out cns.io.crt<br>
* 在namespace: istio-system下创建secret<br>
kubectl create -n istio-system secret tls cns-credential --key=cns.io.key --cert=cns.io.crt<br>
* 在namespace: secns下创建istio ingressgateway<br>
kubectl apply -f istio-gateway.yaml -n secns<br>
* 通过ingressgateway IP访问httpserver服务<br>
curl --resolve httpserver.cns.io:443:$INGRESS_IP https://httpserver.cns.io/healthz -k;echo<br>

```
root@k8snode:~/cloud-native-study/istio# curl --resolve httpserver.cns.io:443:$INGRESS_IP https://httpserver.cns.io/index -k ; echo
<h1>Home page</h1>
root@k8snode:~/cloud-native-study/istio# curl --resolve httpserver.cns.io:443:$INGRESS_IP https://httpserver.cns.io/healthz -k;echo
200
```


## 二、七层路由
* 创建路由规则并测试，/api/healthz重写为/healthz，/api重写为/index
kubectl apply -f istio-r7.yaml -nsecns<br>
curl -H "Host: httpserver.cns.io" $INGRESS_IP/api/healthz;echo<br>
curl -H "Host: httpserver.cns.io" $INGRESS_IP/api;echo<br>

```
root@k8snode:~/cloud-native-study/istio# curl -H "Host: httpserver.cns.io" $INGRESS_IP/api;echo
<h1>Home page</h1>
root@k8snode:~/cloud-native-study/istio# 
root@k8snode:~/cloud-native-study/istio# curl -H "Host: httpserver.cns.io" $INGRESS_IP/api/healthz;echo
200
```


## 三、灰度发布
* 发布V2版本httpserver<br>
kubectl apply -f httpserver-v2.yaml -nsecns<br>
* 创建destinationrule<br>
kubectl apply -f istio-canary.yaml -nsecns<br>
* 创建toolbox供测试使用<br>
kubectl apply -f toolbox.yaml -nsecns<br>
* 进入toolbox测试，v1不带metrics，v2带metrics<br>
TOOL_IP=$(kubectl get pod -nsecns | awk '/toolbox/{print $1}')<br>
kubectl exec -it $TOOL_IP -nsecns -- bash<br>
* 指定headers："user: test"访问v2版本，非user值不是test或未指定headers访问v1版本
curl httpserver/metrics -H 'user: test'<br>
curl httpserver/metrics -H 'user: admin';echo<br>
curl httpserver/metrics;echo<br>
```
[root@toolbox-68f79dd5f8-h8rt5 /]# curl httpserver/metrics -H 'user: test'
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
...

[root@toolbox-68f79dd5f8-h8rt5 /]# curl httpserver/metrics -H 'user: admin';echo
<h1 style="color: red">Page not found.</h1>

[root@toolbox-68f79dd5f8-h8rt5 /]# curl httpserver/metrics;echo
fault filter abort
```


## 四、超时与故障注入
* 进入toolbox测试，v2的index添加了10-2000ms随机延时，当超过1秒时直接超时退出。<br>
kubectl exec -it $INGRESS_IP -nsecns -- bash<br>
* 通过指定headers访问到v2，测试超时退出情况<br>
curl httpserver/index -H 'user: test';echo<br>
```
[root@toolbox-68f79dd5f8-h8rt5 /]# curl httpserver/index -H 'user: test';echo
upstream request timeout
```
* 通过不带headers访问到v1，测试故障注入<br>
curl httpserver/index;echo<br>
```
[root@toolbox-68f79dd5f8-h8rt5 /]# curl httpserver/index;echo
fault filter abort
```


## 五、open tracing 接入
* 安装jaeger日志分析组件<br>
kubectl apply -f jaeger.yaml<br>
* 修改tracing simpling采样比率为100，默认1%<br>
kubectl edit configmap istio -n istio-system<br>

* 创建namespace: tracing并打上istio-injection标签<br>
kubectl create ns tracing<br>
kubectl label ns tracing istio-injection=enabled<br>

* 创建三个服务，请求service0时，它将请求转给service1，service1收到请求转给service2，service2返回结果<br>
kubectl apply -f httpserver0.yaml -n tracing<br>
kubectl apply -f httpserver1.yaml -n tracing<br>
kubectl apply -f httpserver2.yaml -n tracing<br>
* 给service0创建ingressgateway<br>
kubectl apply -f istio-tracing.yaml -n tracing<br>
* 通过ingressgateway service ip访问service0，最终请求被service2处理，可查看pod service2日志验证<br>
curl ${INGRESS_IP}/service0<br>
```
root@k8snode:~/cloud-native-study/istio# curl ${INGRESS_IP}/service0
===================Details of the http request header:============
HTTP/1.1 503 Service Unavailable
Content-Length: 19
Content-Type: text/plain
Date: Mon, 03 Jan 2022 09:06:26 GMT
Server: envoy
...
===================Details of the http request header:============
X-Request-Id=[cd6bef14-99e9-9ada-b7c6-f8f4b7637d33]
X-Forwarded-Client-Cert=[By=spiffe://cluster.local/ns/tracing/sa/default;Hash=53c8ee38bd8a9b8780344045b13efaee2665c3026bbdfa5cdb28eb29cbe8ec1c;Subject="";URI=spiffe://cluster.local/ns/tracing/sa/default]
...
```


* 开启jaeger dashboard，通过浏览器查看整个调用链路信息<br>
istioctl dashboard jaeger --address=0.0.0.0<br>