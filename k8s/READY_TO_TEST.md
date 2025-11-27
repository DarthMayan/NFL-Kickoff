# ✅ Sistema Listo para Load Testing

## Estado Actual

### ✅ Metrics Server: FUNCIONANDO
```bash
kubectl top nodes
# Output: docker-desktop   1597m   19%   2524Mi   32%

kubectl top pods -n kickoff
# Output: Todos los pods mostrando CPU y Memory
```

### ✅ HPA: FUNCIONANDO
```bash
kubectl get hpa -n kickoff
# Output:
# NAME             REFERENCE                       TARGETS       MINPODS   MAXPODS   REPLICAS
# gateway-hpa      Deployment/gateway-service      cpu: 1%/50%   2         10        2
# prediction-hpa   Deployment/prediction-service   cpu: 1%/50%   3         10        3
```

### ✅ Todos los Pods: RUNNING
- Gateway: 2 pods
- User: 2 pods
- Game: 2 pods
- Prediction: 3 pods
- Leaderboard: 2 pods

## 🎯 Gateway Access

**NodePort:** 31859
**URL:** http://localhost:31859

### Endpoints Disponibles:

```bash
# Teams
curl http://localhost:31859/teams

# Users
curl http://localhost:31859/users

# Games
curl http://localhost:31859/games

# Leaderboard
curl http://localhost:31859/leaderboard

# Health check
curl http://localhost:31859/health
```

## 🚀 Ejecutar Load Test

### Opción 1: Load Test Simple (recomendado para empezar)

Abre **3 terminales**:

**Terminal 1 - Monitorear HPA:**
```bash
kubectl get hpa -n kickoff -w
```

**Terminal 2 - Monitorear Pods:**
```bash
kubectl get pods -n kickoff -w
```

**Terminal 3 - Generar Carga:**
```bash
# Generar carga simple con curl en loop
while true; do curl -s http://localhost:31859/teams > nul; done
```

O usa el script (si ya tienes curl instalado):
```bash
cd k8s
load-test.bat
```

### Opción 2: Load Test Agresivo

Para ver el escalamiento más rápido, ejecuta múltiples instancias en paralelo:

```bash
# Terminal 1 y 2: igual que arriba (monitorear HPA y pods)

# Terminal 3: Abrir múltiples procesos de carga
start cmd /c "while true; do curl -s http://localhost:31859/teams > nul; done"
start cmd /c "while true; do curl -s http://localhost:31859/games > nul; done"
start cmd /c "while true; do curl -s http://localhost:31859/users > nul; done"
start cmd /c "while true; do curl -s http://localhost:31859/leaderboard > nul; done"
```

O usa el script agresivo:
```bash
cd k8s
aggressive-load-test.bat
```

## 📊 Qué Observar Durante el Load Test

### 1. CPU Aumentando
En Terminal 1 (HPA), verás:
```
NAME             REFERENCE                       TARGETS        MINPODS   MAXPODS   REPLICAS
gateway-hpa      Deployment/gateway-service      cpu: 1%/50%    2         10        2
```

Luego con carga:
```
gateway-hpa      Deployment/gateway-service      cpu: 45%/50%   2         10        2
gateway-hpa      Deployment/gateway-service      cpu: 65%/50%   2         10        3  ← Escaló!
gateway-hpa      Deployment/gateway-service      cpu: 52%/50%   2         10        4  ← Sigue escalando
```

### 2. Nuevos Pods Creándose
En Terminal 2 (Pods), verás:
```
NAME                                  READY   STATUS              RESTARTS   AGE
gateway-service-5b8fb794cc-pzmcp      1/1     Running             0          8h
gateway-service-5b8fb794cc-w4hhg      1/1     Running             0          8h
gateway-service-5b8fb794cc-xyz12      0/1     ContainerCreating   0          3s  ← Nuevo!
```

Luego:
```
gateway-service-5b8fb794cc-xyz12      1/1     Running             0          15s  ← Listo!
gateway-service-5b8fb794cc-abc34      0/1     ContainerCreating   0          2s   ← Otro nuevo!
```

### 3. Verificar Métricas de CPU
```bash
kubectl top pods -n kickoff
```

Deberías ver CPU aumentando:
```
NAME                                  CPU(cores)   MEMORY(bytes)
gateway-service-5b8fb794cc-pzmcp      45m          8Mi      ← CPU alta
gateway-service-5b8fb794cc-w4hhg      52m          9Mi      ← CPU alta
gateway-service-5b8fb794cc-xyz12      38m          7Mi      ← Nuevo pod ayudando
```

## ⏱️ Timeline Esperado

### Minuto 0-1: Inicio de Carga
- CPU aumenta de 1% → 40-60%
- HPA detecta que estamos sobre el target (50%)

### Minuto 1-2: Primer Escalamiento
- HPA decide escalar
- Nuevos pods comienzan a crearse
- Pods pasan de `Pending` → `ContainerCreating` → `Running`

### Minuto 2-3: Pods Listos
- Nuevos pods están `Running` y recibiendo tráfico
- CPU se distribuye entre más pods
- Si todavía está sobre 50%, HPA escalará más

### Minuto 3-5: Estabilización
- CPU debería estar cerca del target (45-55%)
- REPLICAS muestra el nuevo número (ej: 4-6 pods)
- Sistema estable

### Detener Carga + 1 minuto: Scale Down Comienza
- CPU baja a 5-10%
- HPA espera 60 segundos (stabilizationWindow) antes de scale down
- TARGETS muestra valores bajos

### Detener Carga + 2-3 minutos: Scale Down Ejecutado
- HPA reduce pods gradualmente
- Pods pasan a `Terminating`
- Regresa a minReplicas (2 para gateway, 3 para prediction)

## 🎯 Configuración Actual del HPA

### Gateway Service
```yaml
minReplicas: 2
maxReplicas: 10
target: 50% CPU
scaleUp: Inmediato (duplica pods o +2 cada 15s)
scaleDown: Espera 60s, luego reduce 50% por minuto
```

### Prediction Service
```yaml
minReplicas: 3
maxReplicas: 10
target: 50% CPU
scaleUp: Inmediato (duplica pods o +3 cada 15s)
scaleDown: Espera 60s, luego reduce 50% por minuto
```

**Nota:** Configuración simplificada solo con CPU porque Docker Desktop tiene problemas con las métricas de Memory en HPA. CPU funciona perfectamente.

## 🔧 Comandos Útiles Durante el Test

### Ver detalles de un HPA
```bash
kubectl describe hpa gateway-hpa -n kickoff
```

### Ver logs de un pod específico
```bash
kubectl logs -n kickoff <pod-name> -f
```

### Ver eventos de escalamiento
```bash
kubectl get events -n kickoff --sort-by='.lastTimestamp' | grep -i hpa
```

### Forzar scale manual (para pruebas)
```bash
kubectl scale deployment gateway-service -n kickoff --replicas=5
```

### Resetear a estado inicial
```bash
kubectl scale deployment gateway-service -n kickoff --replicas=2
kubectl scale deployment prediction-service -n kickoff --replicas=3
```

## ✅ Checklist de Verificación

Antes de empezar el load test, verifica:

- [x] Metrics Server instalado y funcionando
- [x] `kubectl top nodes` funciona
- [x] `kubectl top pods -n kickoff` funciona
- [x] HPA muestra `cpu: X%/50%` (no `<unknown>`)
- [x] Todos los pods en estado `Running`
- [x] Gateway accesible en http://localhost:31859
- [x] curl http://localhost:31859/teams responde correctamente

## 🎉 ¡Listo para Probar!

Tu sistema está completamente funcional y listo para demostrar autoscaling con Kubernetes HPA.

**Recomendación:** Empieza con el load test simple (Terminal 3 con un solo curl loop) para ver cómo funciona, luego prueba el agresivo para ver escalamiento más dramático.

**Tip:** Graba la pantalla o toma screenshots de las terminales mostrando el escalamiento en acción - es muy visual y demuestra que todo funciona correctamente.

## 📝 Nota sobre LoadBalancer

El `EXTERNAL-IP` del gateway-service muestra `<pending>` porque Docker Desktop no provee IPs externas automáticamente. Esto es normal. Usamos el NodePort (31859) para acceder al servicio desde localhost.

Si quieres que el LoadBalancer funcione, puedes:
```bash
kubectl port-forward -n kickoff svc/gateway-service 8080:8080
```

Luego acceder en http://localhost:8080
