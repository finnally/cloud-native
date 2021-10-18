#!/bin/bash

container_name=$(awk '{print $1}' /tmp/.info)
image_name=$(awk '{print $2}' /tmp/.info)

echo "remove container ${container_name} and image ${image_name}"
kill $pid && sleep 2 && docker rm ${container_name}
docker rmi ${image_name}
