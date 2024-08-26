docker build -t dockerstats -f Dockerfile .

docker run --name lls -p 9099:9099 -v "/var/run/docker.sock:/var/run/docker.sock" dockerstats


docker stop lls

docker rm lls

docker rmi dockerstats

docker run -d \
--restart always \
--net=bridge \
--env=TZ=Asia/Shanghai \
--volume=/:/rootfs:ro \
--volume=/var/run:/var/run:ro \
--volume=/sys:/sys:ro \
--volume=/var/lib/docker/:/var/lib/docker:ro \
--volume=/dev/disk/:/dev/disk:ro \
--publish=9222:8080 \
--detach=true \
--name=cadvisor_exporter \
--privileged \
--device=/dev/kmsg \
registry.ap-southeast-1.aliyuncs.com/ak_system/cadvisor_exporter:latest \
-enable_metrics cpu,memory,network