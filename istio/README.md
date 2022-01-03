# istio发布httpserver服务

## 一、httpserver 服务以 Istio Ingress Gateway 的形式发布
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
kubectl apply -f istio-spec.yaml -n secns<br>
* 通过ingressgateway IP访问httpserver服务<br>
curl --resolve httpserver.cns.io:443:$(kubectl get svc -nistio-system | awk '/istio-ingressgateway/{print $3}') https://httpserver.cns.io/healthz -k<br>


## 二、七层路由
kubectl apply -f istio-r7.yaml -nsecns<br>
curl --resolve httpserver.cns.io:443:10.99.93.117 https://httpserver.cns.io/api/healthz -k<br>
curl --resolve httpserver.cns.io:443:10.99.93.117 https://httpserver.cns.io/api -k<br>

## 三、灰度发布
* 发布V2版本httpserver<br>
kubectl apply -f httpserver-v2.yaml -nsecns<br>
* 创建destinationrule<br>
kubectl apply -f istio-canary.yaml -nsecns<br>
* 创建toolbox供测试使用<br>
kubectl apply -f toolbox.yaml -nsecns<br>
* 进入toolbox测试，v1不带metrics，v2带metrics<br>
kubectl exec -it $(kubectl get pod -nsecns|awk '/toolbox/{print $1}') -nsecns -- bash<br>
curl httpserver/metrics -H 'user: test'<br>


## 四、超时与故障注入
* 进入toolbox测试，v2的index添加了10-2000ms随机延时，当超过1秒时直接超时退出。<br>
kubectl exec -it $(kubectl get pod -nsecns|awk '/toolbox/{print $1}') -nsecns -- bash<br>
* 通过指定headers访问到v2，测试超时退出情况<br>
curl httpserver/index -H 'user: test'<br>
* 通过不带headers访问到v1，测试故障注入<br>
curl httpserver/index -H<br>


## 五、open tracing 接入
* 安装jaeger日志分析组件<br>
kubectl apply -f jaeger.yaml<br>
* 修改tracing simpling采样比率为100，默认1%<br>
kubectl edit configmap istio -n istio-system<br>

* 创建namespace: tracing并打上istio-injection标签<br>
kubectl create ns tracing<br>
kubectl label ns tracing istio-injection=enabled<br>

* 创建三个服务，请求service0时，它将请求转给service1，service1收到请求转给service2，service2返回结果<br>
kubectl apply -f service0.yaml -n tracing<br>
kubectl apply -f service1.yaml -n tracing<br>
kubectl apply -f service2.yaml -n tracing<br>
* 给service0创建ingressgateway<br>
kubectl apply -f istio-tracing.yaml -n tracing<br>
* 通过ingressgateway service ip访问service0，最终请求被service2处理，可查看pod service2日志验证<br>
curl $(kubectl get svc -nistio-system | awk '/istio-ingressgateway/{print $3}')/service0<br>

* 开启jaeger dashboard，通过浏览器查看整个调用链路信息<br>
istioctl dashboard jaeger --address=0.0.0.0<br>