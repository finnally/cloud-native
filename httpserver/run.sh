#!/bin/bash 

container_name="httpserver-$RANDOM"
image_name="hellionc/httpserver:v1.0"

color_print () {
    case $1 in
        info)
        printf '\033[1;32;40m%b\033[0m\n' "$2"
        ;;
        failed)
        printf '\033[1;32;40m%b\033[0m\n' "$2"
        ;;
    esac
}

run () {
    echo ${container_name} ${image_name} > /tmp/.info
    color_print info "start docker container ${container_name} ..."
    if ! which docker >/dev/null 2>&1;then
        color_print failed "docker command not found."
        exit 404
    elif ! docker ps -a >/dev/null 2>&1;then
        color_print failed "docker serivice exception."
        exit 500
    fi
    
    docker run -d --name=${container_name} -P ${image_name}

    color_print info "view ${container_name} ip configuration."
    pid=$(docker inspect -f {{.State.Pid}} ${container_name} 2>/dev/null)
    if [ "$pid" -eq 0 ];then
        color_print failed "container ${container_name} run failed."
    fi
    if ! which nsenter >/dev/null 2>&1;then
        color_print failed "nsenter command not found."
        exit 404
    fi
    nsenter -t $pid -n ip a
}

make -s push
run
