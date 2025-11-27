@echo off
echo ========================================
echo Removing Kickoff NFL from Kubernetes
echo ========================================
echo.

echo Deleting all resources in kickoff namespace...
kubectl delete namespace kickoff

echo.
echo ========================================
echo Cleanup Complete!
echo ========================================
echo.
pause
