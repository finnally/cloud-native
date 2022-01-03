# istio发布httpserver服务

## httpserver 服务以 Istio Ingress Gateway 的形式发布:
* 创建namespace: secns
kubectl create ns secns
* 给该ns打上istio-injection标签，使创建pod时自动注入istio-proxy sidecar
kubectl label ns secns istio-injection=enabled
* 在namespace: secns下创建pod和service
kubectl create -f httpserver.yaml -n secns
* 生成SSL证书
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=cns Inc./CN=*.cns.io' -keyout cns.io.key -out cns.io.crt
* 在namespace: istio-system下创建secret
kubectl create -n istio-system secret tls cns-credential --key=cns.io.key --cert=cns.io.crt
* 在namespace: secns下创建istio ingressgateway
kubectl apply -f istio-spec.yaml -n secns
* 通过ingressgateway IP访问httpserver服务
curl --resolve httpserver.cns.io:443:$(kubectl get svc -nistio-system | awk '/istio-ingressgateway/{print $3}') https://httpserver.cns.io/healthz -k


## 七层路由
kubectl apply -f istio-r7.yaml -nsecns
curl --resolve httpserver.cns.io:443:10.99.93.117 https://httpserver.cns.io/api/healthz -k
curl --resolve httpserver.cns.io:443:10.99.93.117 https://httpserver.cns.io/api -k

## 灰度发布
* 发布V2版本httpserver
kubectl apply -f httpserver-v2.yaml -nsecns
* 创建destinationrule
kubectl apply -f istio-canary.yaml -nsecns
* 创建toolbox供测试使用
kubectl apply -f toolbox.yaml -nsecns
* 进入toolbox测试，v1不带metrics，v2带metrics
kubectl exec -it $(kubectl get pod -nsecns|awk '/toolbox/{print $1}') -nsecns -- bash
curl httpserver/metrics -H 'user: test'


## 超时与故障注入
* 进入toolbox测试，v2的index添加了10-2000ms随机延时，当超过1秒时直接超时退出。
kubectl exec -it $(kubectl get pod -nsecns|awk '/toolbox/{print $1}') -nsecns -- bash
* 通过指定headers访问到v2，测试超时退出情况
curl httpserver/index -H 'user: test'
* 通过不带headers访问到v1，测试故障注入
curl httpserver/index -H


## open tracing 接入
* 安装jaeger日志分析组件
kubectl apply -f jaeger.yaml
* 修改tracing simpling采样比率为100，默认1%
kubectl edit configmap istio -n istio-system

* 创建namespace: tracing并打上istio-injection标签
kubectl create ns tracing
kubectl label ns tracing istio-injection=enabled

* 创建三个服务，请求service0时，它将请求转给service1，service1收到请求转给service2，service2返回结果
kubectl apply -f service0.yaml -n tracing
kubectl apply -f service1.yaml -n tracing
kubectl apply -f service2.yaml -n tracing
* 给service0创建ingressgateway
kubectl apply -f istio-tracing.yaml -n tracing
* 通过ingressgateway service ip访问service0，最终请求被service2处理，可查看pod service2日志验证
curl $(kubectl get svc -nistio-system | awk '/istio-ingressgateway/{print $3}')/service0

* 开启jaeger dashboard，通过浏览器查看整个调用链路信息
istioctl dashboard jaeger --address=0.0.0.0