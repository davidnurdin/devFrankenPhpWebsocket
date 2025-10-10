#!/bin/bash

echo "Test de la correction du crash zend_mm_heap corrupted"
echo "=================================================="

# Compilation
echo "1. Compilation du module..."
cd /home/david/poc/frankenphpWebsocket/frankenphp-websocket
if go build -v; then
    echo "✅ Compilation réussie"
else
    echo "❌ Échec de la compilation"
    exit 1
fi

echo ""
echo "2. Test avec le worker PHP..."
cd /home/david/poc/frankenphpWebsocket

# Test rapide pour vérifier que le worker peut démarrer
timeout 5s php websocket-worker.php &
WORKER_PID=$!

sleep 2

# Vérifier si le processus est encore en vie
if kill -0 $WORKER_PID 2>/dev/null; then
    echo "✅ Worker PHP démarré sans crash"
    kill $WORKER_PID 2>/dev/null
else
    echo "❌ Worker PHP a crashé"
    exit 1
fi

echo ""
echo "3. Résumé des modifications apportées :"
echo "   - Suppression de la variable globale frankenphp_ws_current_retval"
echo "   - Passage du zval* directement en paramètre"
echo "   - Protection par mutex contre les appels concurrents"
echo "   - Élimination des race conditions entre threads"

echo ""
echo "✅ Test terminé - La correction devrait résoudre le crash zend_mm_heap corrupted"
