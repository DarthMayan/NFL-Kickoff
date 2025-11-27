/**
 * K6 Stress Test Script - Kickoff NFL
 *
 * Este script realiza un test de estrés con rampa de carga
 * para determinar la capacidad máxima del cluster.
 *
 * Uso:
 *   k6 run k6-stress-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';

// Métricas personalizadas
const errorRate = new Rate('errors');
const successfulRequests = new Counter('successful_requests');
const requestDuration = new Trend('custom_request_duration');

// Configuración del stress test
export const options = {
    scenarios: {
        stress_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 10 },   // Calentar: 10 usuarios en 1 min
                { duration: '2m', target: 30 },   // Escalar: 30 usuarios en 2 min
                { duration: '2m', target: 50 },   // Escalar: 50 usuarios en 2 min
                { duration: '2m', target: 100 },  // Estrés: 100 usuarios en 2 min
                { duration: '3m', target: 100 },  // Sostener: 100 usuarios por 3 min
                { duration: '2m', target: 50 },   // Bajar: 50 usuarios en 2 min
                { duration: '1m', target: 0 },    // Enfriar: 0 usuarios en 1 min
            ],
        },
    },

    thresholds: {
        'http_req_duration': [
            'p(50)<200',    // 50% de requests < 200ms
            'p(95)<1000',   // 95% de requests < 1s
            'p(99)<2000',   // 99% de requests < 2s
        ],
        'http_req_failed': ['rate<0.1'],  // Menos de 10% de errores
        'errors': ['rate<0.1'],
    },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    const endpoint = selectRandomEndpoint();
    const startTime = Date.now();

    const response = http.get(`${BASE_URL}${endpoint}`, {
        timeout: '10s',
    });

    const duration = Date.now() - startTime;
    requestDuration.add(duration);

    const result = check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 2000ms': (r) => r.timings.duration < 2000,
        'has content': (r) => r.body.length > 0,
    });

    if (result) {
        successfulRequests.add(1);
    } else {
        errorRate.add(1);
    }

    // Simular comportamiento de usuario (500ms - 2s)
    sleep(Math.random() * 1.5 + 0.5);
}

function selectRandomEndpoint() {
    const endpoints = [
        '/api/teams',
        '/api/users',
        '/api/games',
        '/api/leaderboard',
        '/health',
    ];
    return endpoints[Math.floor(Math.random() * endpoints.length)];
}

export function setup() {
    console.log('\n' + '='.repeat(70));
    console.log(' K6 STRESS TEST - Kickoff NFL Kubernetes Cluster');
    console.log('='.repeat(70));
    console.log('');
    console.log('Este test determinará la capacidad máxima del cluster:');
    console.log('  • Fase 1: Calentar con 10 usuarios');
    console.log('  • Fase 2: Escalar a 30 usuarios');
    console.log('  • Fase 3: Escalar a 50 usuarios');
    console.log('  • Fase 4: Estrés con 100 usuarios');
    console.log('  • Fase 5: Sostener 100 usuarios por 3 minutos');
    console.log('  • Fase 6: Reducir y enfriar');
    console.log('');
    console.log('Duración total: ~13 minutos');
    console.log('');

    // Verificar gateway
    const healthCheck = http.get(`${BASE_URL}/health`);
    if (healthCheck.status !== 200) {
        console.error('❌ ERROR: Gateway no disponible!');
        console.error('   Ejecuta: kubectl port-forward -n kickoff svc/gateway-service 8080:8080');
        throw new Error('Gateway no disponible');
    }

    console.log('✓ Gateway disponible y listo');
    console.log('');
    console.log('Monitorea el HPA en otra terminal:');
    console.log('  kubectl get hpa -n kickoff -w');
    console.log('');
    console.log('Iniciando en 3 segundos...');
    console.log('='.repeat(70) + '\n');

    return { startTime: new Date() };
}

export function teardown(data) {
    const duration = (new Date() - data.startTime) / 1000;
    const minutes = Math.floor(duration / 60);
    const seconds = Math.floor(duration % 60);

    console.log('\n' + '='.repeat(70));
    console.log(' STRESS TEST COMPLETADO');
    console.log('='.repeat(70));
    console.log('');
    console.log(`Duración total: ${minutes}m ${seconds}s`);
    console.log('');
    console.log('Revisa los resultados arriba para:');
    console.log('  • Total de requests ejecutados');
    console.log('  • Requests por segundo (RPS)');
    console.log('  • Percentiles de latencia (p50, p95, p99)');
    console.log('  • Tasa de error');
    console.log('');
    console.log('Verifica el estado del HPA:');
    console.log('  kubectl get hpa -n kickoff');
    console.log('');
    console.log('Número de pods escalados:');
    console.log('  kubectl get pods -n kickoff');
    console.log('');
    console.log('='.repeat(70) + '\n');
}
