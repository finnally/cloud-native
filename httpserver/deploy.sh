#!/bin/bash

yamlConfig="\
	httpserver-deployment.yaml \
	ingress-deployment.yaml \
	secret.yaml \
	ingress.yaml"
    
deploy() {
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=hellion.com/O=hellion"
    kubectl create secret tls hellion-tls --cert=tls.crt --key=tls.key --dry-run=client -oyaml > secret.yaml
    rm -f tls.crt tls.key
    
    for yaml in $yamlConfig
    do
	if [ "${yaml}" == "ingress.yaml" ];then
            while :
	    do
		controllerName=$(kubectl get pod -n ingress-nginx | awk '/controller/{print $1}')
		controllerStatus=$(kubectl get pod -n ingress-nginx | awk '/controller/{print $3}')
		if [ ${controllerStatus} == "Running" ];then
		    break
		fi
		echo "Wait for ingress-nginx-controller running..."
		sleep 1
	    done
	fi
	sleep 10
        kubectl apply -f $yaml
    done
}

accessTest() {
    ingressServiceIp=$(kubectl get svc -n ingress-nginx | awk '/LoadBalancer/{print $3}')
    nodePort=$(kubectl get svc -n ingress-nginx | awk -F'[ /:]+' '/LoadBalancer/{print $8}')
    echo "access from ingress service clusterIp"
    echo "curl https://${ingressServiceIp}/index"
    curl -H "Host: hellion.com" https://${ingressServiceIp}/index -k ; echo ; echo
    
    echo "access from hostname"
    echo "curl https://$(hostname):${nodePort}/index"
    curl -H "Host: hellion.com" https://$(hostname):${nodePort}/index -k ; echo
}

logrotateConfig() {
    containerId=$(docker ps | awk '/httpserver/&&!/pause/{print $1}')
    touch /etc/logrotate.d/httpserver && >/etc/logrotate.d/httpserver
    for id in ${containerId}
    do
        logPath=$(docker inspect -f {{.LogPath}} $id)
        cat >> /etc/logrotate.d/httpserver <<EOF
${logPath} {
	create 644 root root
	notifempty
	missingok
	copytruncate
	noolddir
	rotate 2
	size=10M
}

EOF
    done
}

clean() {
    for yaml in $yamlConfig
    do
        kubectl delete -f $yaml
    done
}

if [ -z $1 ];then
    echo "Usage: bash $0 deploy|accessTest|logrotateConfig|clean"
fi

cd manifests
case $1 in
deploy)
    deploy
    ;;
accessTest)
    accessTest
    ;;
logrotateConfig)
    logrotateConfig
    ;;
clean)
    clean
    ;;
esac
