<?php

// Test de la nouvelle fonctionnalité de ping périodique

echo "=== Test du ping périodique ===\n";

// Simuler une connexion WebSocket
$connectionId = "test_connection_123";

echo "1. Activation du ping avec intervalle de 5 secondes (5000ms)\n";
$result = frankenphp_ws_enablePing($connectionId, 5000);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n2. Attente de 12 secondes pour voir les pings périodiques...\n";
sleep(12);

echo "\n3. Vérification du temps de ping\n";
$pingTime = frankenphp_ws_getClientPingTime($connectionId);
echo "Temps de ping: " . $pingTime . " nanosecondes (" . ($pingTime / 1000000) . " ms)\n";

echo "\n4. Désactivation du ping périodique\n";
$result = frankenphp_ws_disablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n5. Test avec intervalle par défaut (pas de ping périodique)\n";
$result = frankenphp_ws_enablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n6. Désactivation finale\n";
$result = frankenphp_ws_disablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n=== Test terminé ===\n";
