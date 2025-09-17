@echo off
echo Stopping Consul dev server...
docker stop dev-consul
docker rm dev-consul
echo Consul stopped and removed