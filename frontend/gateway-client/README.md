Kickoff - NFL Predictions App

Frontend tipo dashboard para visualizar datos del Gateway en tiempo real.

Cómo ejecutar:

**Python (recomendado):**
```powershell
cd frontend/gateway-client
python -m http.server 5500
```

**Node.js (alternativa):**
```bash
cd frontend/gateway-client
npx serve -l 5500
```

Luego abre en tu navegador: **http://localhost:5500**

Funcionalidades:
- Dashboard con secciones: Teams, Games, Users, Leaderboard, Predictions
- Muestra data en tiempo real del Gateway
- Auto-refresh cada 30 segundos
- Responsive design (funciona en móvil)

Requisitos:
- Gateway accesible en `http://localhost:8080`
- Si usas Kind/Kubernetes sin exposición directa, haz port-forward:
  ```powershell
  kubectl port-forward -n kickoff svc/gateway-service 8080:8080
  ```
