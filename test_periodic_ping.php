<?php

// Test de la nouvelle fonctionnalité de ping périodique avec traces détaillées

echo "=== Test du ping périodique avec traces ===\n";

// Simuler une connexion WebSocket
$connectionId = "test_connection_123";

echo "1. Activation du ping avec intervalle de 3 secondes (3000ms)\n";
$result = frankenphp_ws_enablePing($connectionId, 3000);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n2. Attente de 10 secondes pour voir les pings périodiques...\n";
for ($i = 1; $i <= 10; $i++) {
    sleep(1);
    echo "Seconde $i/10...\n";
    
    // Vérifier le temps de ping toutes les 2 secondes
    if ($i % 2 == 0) {
        $pingTime = frankenphp_ws_getClientPingTime($connectionId);
        echo "  -> Temps de ping: " . $pingTime . " ns (" . ($pingTime / 1000000) . " ms)\n";
    }
}

echo "\n3. Vérification finale du temps de ping\n";
$pingTime = frankenphp_ws_getClientPingTime($connectionId);
echo "Temps de ping final: " . $pingTime . " nanosecondes (" . ($pingTime / 1000000) . " ms)\n";

echo "\n4. Désactivation du ping périodique\n";
$result = frankenphp_ws_disablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n5. Test avec intervalle par défaut (ping manuel unique)\n";
$result = frankenphp_ws_enablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n6. Attente de 2 secondes pour voir le pong du ping manuel...\n";
sleep(2);

echo "\n7. Vérification après activation manuelle\n";
$pingTime = frankenphp_ws_getClientPingTime($connectionId);
echo "Temps de ping (manuel): " . $pingTime . " nanosecondes (" . ($pingTime / 1000000) . " ms)\n";

echo "\n8. Test d'un autre ping manuel\n";
$result = frankenphp_ws_enablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";
sleep(2);
$pingTime = frankenphp_ws_getClientPingTime($connectionId);
echo "Temps de ping (manuel 2): " . $pingTime . " nanosecondes (" . ($pingTime / 1000000) . " ms)\n";

echo "\n9. Désactivation finale\n";
$result = frankenphp_ws_disablePing($connectionId);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n=== Test terminé ===\n";
echo "Consultez les logs Caddy pour voir les traces détaillées du ping/pong.\n";
