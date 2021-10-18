#!/bin/bash

container_name=$(awk '{print $1}' /tmp/.info)
image_name=$(awk '{print $2}' /tmp/.info)

echo -e "\033[1;32mremove container ${container_name} and image ${image_name}\033[1;0m"
pid=$(docker inspect -f {{.State.Pid}} ${container_name} 2>/dev/null)
kill $pid && sleep 2 && docker rm ${container_name}
docker rmi ${image_name}
