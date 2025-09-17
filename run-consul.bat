@echo off
echo Starting Consul dev server...
docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul consul:1.15.4 agent -server -ui -node=kickoff-server -bootstrap-expect=1 -client=0.0.0.0 -datacenter=kickoff
echo Consul started on http://localhost:8500