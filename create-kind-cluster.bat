@echo off
echo Creating Kind cluster...
kind create cluster --name kickoff --config kind-config.yaml
echo Done!
pause
