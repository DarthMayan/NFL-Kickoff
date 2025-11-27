# Load Testing Guide - Kickoff NFL

Este documento describe cómo realizar pruebas de carga para verificar el funcionamiento del Horizontal Pod Autoscaler (HPA) en Kubernetes.

## Requisitos Previos

1. **Metrics Server**: El HPA requiere Metrics Server para funcionar
2. **Deployment activo**: Todos los servicios deben estar corriendo
3. **curl**: Para generar tráfico HTTP

## Verificar Metrics Server

Antes de comenzar, verifica que Metrics Server esté instalado:

```bash
kubectl get deployment metrics-server -n kube-system
```

Si no está instalado, puedes instalarlo con:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

Para Minikube, habilita el addon:

```bash
minikube addons enable metrics-server
```

## Scripts de Load Testing

### 1. `load-test.bat` - Load Test Moderado

Script básico que genera tráfico secuencial a diferentes endpoints:

```bash
cd k8s
load-test.bat
```

**Características:**
- Rota entre 5 endpoints diferentes
- Muestra el tiempo de respuesta de cada request
- Load moderado para observar el comportamiento gradual del HPA

**Uso:**
- Observa cómo el HPA escala gradualmente
- Ideal para demos y verificación inicial

### 2. `aggressive-load-test.bat` - Load Test Agresivo

Script que genera carga pesada con múltiples procesos paralelos:

```bash
cd k8s
aggressive-load-test.bat
```

**Características:**
- Genera 10 procesos paralelos
- Cada proceso hace 10,000 requests
- Monitorea automáticamente el HPA cada 5 segundos

**Uso:**
- Para verificar que el HPA realmente funciona bajo carga
- Debería ver escalamiento rápido de pods

## Monitoreo Durante Load Testing

### Ver estado de HPA

```bash
kubectl get hpa -n kickoff
```

**Output esperado:**
```
NAME              REFERENCE                    TARGETS          MINPODS  MAXPODS  REPLICAS
gateway-hpa       Deployment/gateway-service   45%/70%, 60%/80% 2        10       4
prediction-hpa    Deployment/prediction-service 65%/60%, 70%/75% 3        15       8
```

Observa:
- **TARGETS**: Métricas actuales vs targets (CPU/Memory)
- **REPLICAS**: Número actual de pods (aumenta durante load)

### Ver pods en tiempo real

```bash
kubectl get pods -n kickoff -w
```

Verás pods nuevos en estado:
1. `Pending` - Siendo creados
2. `ContainerCreating` - Descargando imagen y iniciando
3. `Running` - Listos para servir tráfico

### Ver logs de un pod específico

```bash
kubectl logs -n kickoff <pod-name> --tail=50 -f
```

### Ver métricas de recursos

```bash
kubectl top pods -n kickoff
kubectl top nodes
```

## Configuración HPA

### Gateway Service (más conservador)
```yaml
minReplicas: 2
maxReplicas: 10
metrics:
  - CPU: 70%
  - Memory: 80%
```

### Prediction Service (más agresivo)
```yaml
minReplicas: 3
maxReplicas: 15
metrics:
  - CPU: 60%  # Escala más rápido
  - Memory: 75%
```

### Comportamiento de Escalamiento

**Scale Up (más rápido):**
- Se evalúa cada 15 segundos
- Escala hasta 100% (duplica pods) o +4 pods
- Sin estabilización - responde inmediatamente

**Scale Down (más lento):**
- Espera 5 minutos de estabilidad
- Reduce máximo 50% de pods por minuto
- Previene "flapping" (escalar arriba/abajo repetidamente)

## Experimentos Sugeridos

### Experimento 1: Escalamiento Gradual

1. Iniciar `load-test.bat` (moderado)
2. Observar HPA:
   ```bash
   watch kubectl get hpa -n kickoff
   ```
3. Esperar a ver CPU aumentar gradualmente
4. Observar cuando HPA decide escalar (>70% CPU)
5. Contar cuánto tiempo toma crear nuevos pods

**Resultado esperado:** 2-3 pods adicionales después de 1-2 minutos

### Experimento 2: Escalamiento Agresivo

1. Iniciar `aggressive-load-test.bat`
2. En otra terminal:
   ```bash
   kubectl get pods -n kickoff -w
   ```
3. Observar pods creándose rápidamente
4. Ver el límite máximo (10 para gateway, 15 para prediction)

**Resultado esperado:** Alcanzar maxReplicas en 2-3 minutos

### Experimento 3: Scale Down

1. Detener el load test (Ctrl+C)
2. Observar que el HPA detecta baja carga
3. Esperar 5 minutos (stabilizationWindow)
4. Ver pods siendo terminados gradualmente

**Resultado esperado:** Regreso a minReplicas después de 6-8 minutos

## Verificar que el HPA está funcionando

### Señales de que funciona correctamente:

1. **TARGETS muestra valores**:
   ```
   TARGETS: 45%/70%, 60%/80%  ✅
   TARGETS: <unknown>/<unknown>  ❌ (Metrics Server no está funcionando)
   ```

2. **REPLICAS aumenta durante carga**:
   ```
   REPLICAS: 2 -> 3 -> 5 -> 8  ✅
   REPLICAS: 2 (sin cambios)   ❌
   ```

3. **Events en HPA**:
   ```bash
   kubectl describe hpa gateway-hpa -n kickoff
   ```

   Busca eventos como:
   ```
   ScaledUp: New size: 4; reason: cpu resource utilization above target
   ```

## Troubleshooting

### HPA muestra `<unknown>` en TARGETS

**Problema:** Metrics Server no está funcionando

**Solución:**
```bash
# Verificar Metrics Server
kubectl get pods -n kube-system | grep metrics

# Reinstalar si es necesario
kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### HPA no escala a pesar de alta carga

**Posibles causas:**

1. **Resource requests no definidos**: HPA necesita `resources.requests` en los Deployments
   ```bash
   kubectl get deployment gateway-service -n kickoff -o yaml | grep -A 5 resources
   ```

2. **Carga no suficiente**: Usa `aggressive-load-test.bat`

3. **Tiempo de estabilización**: Espera al menos 15 segundos entre mediciones

### LoadBalancer muestra `<pending>`

En entornos locales (Minikube, Kind), LoadBalancer no recibe IP externa automáticamente.

**Solución - Minikube:**
```bash
minikube tunnel
```

**Solución - Usar NodePort:**
```bash
kubectl port-forward -n kickoff svc/gateway-service 8080:8080
```

## Ejemplo de Sesión Completa

```bash
# Terminal 1: Monitorear HPA
kubectl get hpa -n kickoff -w

# Terminal 2: Monitorear pods
kubectl get pods -n kickoff -w

# Terminal 3: Generar carga
cd k8s
aggressive-load-test.bat

# Observar:
# - HPA TARGETS aumentando (20% -> 50% -> 80%)
# - REPLICAS aumentando (2 -> 4 -> 8)
# - Nuevos pods apareciendo en estado Running
# - CPU/Memory metrics aumentando

# Después de 2-3 minutos, detener carga (Ctrl+C en Terminal 3)

# Observar scale down:
# - TARGETS bajando (80% -> 50% -> 20%)
# - Esperar 5 minutos
# - REPLICAS disminuyendo gradualmente (8 -> 6 -> 4 -> 2)
```

## Métricas Clave a Observar

| Métrica | Normal | Bajo Carga | Escalado |
|---------|--------|------------|----------|
| CPU % | 10-30% | 70-100% | 40-60% |
| Memory % | 20-40% | 75-90% | 50-70% |
| Replicas Gateway | 2 | 2-4 | 6-10 |
| Replicas Prediction | 3 | 3-8 | 10-15 |
| Response Time | <100ms | 200-500ms | <150ms |

## Conclusión

El HPA está funcionando correctamente cuando:
- ✅ Responde a aumentos de carga en <30 segundos
- ✅ Escala hasta maxReplicas bajo carga sostenida
- ✅ Scale down gradualmente después de 5-6 minutos sin carga
- ✅ Mantiene las métricas dentro de los targets configurados
- ✅ Response times se mantienen razonables incluso bajo carga
