/**
 * K6 Load Test Script para Kickoff NFL
 *
 * Este script simula usuarios concurrentes accediendo a diferentes endpoints
 * del sistema de predicciones NFL para medir la capacidad del cluster.
 *
 * Uso:
 *   k6 run k6-load-test.js                    # Test básico
 *   k6 run --vus 10 --duration 30s k6-load-test.js  # 10 usuarios por 30s
 *   k6 run --vus 50 --duration 1m k6-load-test.js   # 50 usuarios por 1 min
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Métricas personalizadas
const errorRate = new Rate('errors');

// Configuración del test
export const options = {
    // Escenarios de carga
    scenarios: {
        // Escenario 1: Carga constante
        constant_load: {
            executor: 'constant-vus',
            vus: 20,                    // 20 usuarios virtuales
            duration: '1m',             // Durante 1 minuto
        },

        // Escenario 2: Rampa de carga (comentado por defecto)
        // ramp_load: {
        //     executor: 'ramping-vus',
        //     startVUs: 0,
        //     stages: [
        //         { duration: '30s', target: 10 },  // Subir a 10 usuarios en 30s
        //         { duration: '1m', target: 20 },   // Subir a 20 usuarios en 1m
        //         { duration: '30s', target: 50 },  // Subir a 50 usuarios en 30s
        //         { duration: '1m', target: 50 },   // Mantener 50 usuarios por 1m
        //         { duration: '30s', target: 0 },   // Bajar a 0 usuarios en 30s
        //     ],
        // },
    },

    // Thresholds - Criterios de éxito/fallo
    thresholds: {
        'http_req_duration': ['p(95)<500'],     // 95% de requests < 500ms
        'http_req_failed': ['rate<0.05'],       // Menos de 5% de errores
        'errors': ['rate<0.05'],                // Menos de 5% de errores de validación
    },
};

// URL base del gateway (debe estar con port-forward activo)
const BASE_URL = 'http://localhost:8080';

// Función principal que ejecuta cada usuario virtual
export default function () {
    // Seleccionar endpoint aleatorio para simular comportamiento real
    const endpoint = selectRandomEndpoint();

    // Ejecutar request
    const response = http.get(`${BASE_URL}${endpoint}`);

    // Validar respuesta
    const result = check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
        'has valid JSON': (r) => {
            try {
                JSON.parse(r.body);
                return true;
            } catch (e) {
                return false;
            }
        },
    });

    // Registrar errores
    errorRate.add(!result);

    // Simular tiempo de "pensar" del usuario (1-3 segundos)
    sleep(Math.random() * 2 + 1);
}

// Función para seleccionar endpoint aleatorio
function selectRandomEndpoint() {
    const endpoints = [
        '/api/teams',           // Listar equipos
        '/api/users',           // Listar usuarios
        '/api/games',           // Listar juegos
        '/api/leaderboard',     // Ver leaderboard
        '/health',              // Health check
    ];

    return endpoints[Math.floor(Math.random() * endpoints.length)];
}

// Función que se ejecuta al inicio del test (una vez)
export function setup() {
    console.log('='.repeat(60));
    console.log('Iniciando K6 Load Test - Kickoff NFL');
    console.log('='.repeat(60));
    console.log(`Target: ${BASE_URL}`);
    console.log('');

    // Verificar que el gateway está disponible
    const response = http.get(`${BASE_URL}/health`);
    if (response.status !== 200) {
        console.error('ERROR: Gateway no está disponible!');
        console.error('Asegúrate de ejecutar: kubectl port-forward -n kickoff svc/gateway-service 8080:8080');
        throw new Error('Gateway no disponible');
    }

    console.log('✓ Gateway disponible');
    console.log('');

    return { startTime: new Date() };
}

// Función que se ejecuta al final del test (una vez)
export function teardown(data) {
    const duration = (new Date() - data.startTime) / 1000;
    console.log('');
    console.log('='.repeat(60));
    console.log('Test Completado');
    console.log('='.repeat(60));
    console.log(`Duración total: ${duration.toFixed(2)} segundos`);
    console.log('');
    console.log('Para ver el HPA durante el test:');
    console.log('  kubectl get hpa -n kickoff -w');
    console.log('');
}
