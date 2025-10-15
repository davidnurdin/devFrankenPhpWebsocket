<?php

// Test de la fonctionnalité Queue Counter

echo "=== Test de la Queue Counter ===\n";

// Simuler une connexion WebSocket
$connectionId = "test_connection_456";

echo "1. Activation de la queue counter (50 messages max, 30 minutes max)\n";
$result = frankenphp_ws_enableQueueCounter($connectionId, 50, 1800);
echo "Résultat: " . ($result ? "SUCCESS" : "FAILED") . "\n";

echo "\n2. Vérification du compteur initial\n";
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur initial: " . $counter . "\n";

echo "\n3. Envoi de plusieurs messages (ils seront automatiquement tracés)\n";
for ($i = 1; $i <= 5; $i++) {
    frankenphp_ws_send($connectionId, "Message de test $i", "/test");
    echo "Message $i envoyé\n";
    sleep(1);
}

echo "\n4. Vérification du compteur après envoi\n";
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur après envoi: " . $counter . "\n";

echo "\n5. Consultation de la queue des messages\n";
$queue = frankenphp_ws_getClientMessageQueue($connectionId);
echo "Nombre de messages en queue: " . count($queue) . "\n";

foreach ($queue as $index => $messageData) {
    echo "Message " . ($index + 1) . ": " . $messageData . "\n";
}

echo "\n6. Test avec sendAll (sera tracé pour ce client)\n";
$sentCount = frankenphp_ws_sendAll("Message pour tous les clients", "/broadcast");
echo "Messages envoyés en masse: " . $sentCount . "\n";

echo "\n7. Vérification du compteur après sendAll\n";
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur après sendAll: " . $counter . "\n";

echo "\n8. Consultation de la queue mise à jour\n";
$queue = frankenphp_ws_getClientMessageQueue($connectionId);
echo "Nombre de messages en queue: " . count($queue) . "\n";

echo "\n9. Test de vidage de la queue\n";
$clearResult = frankenphp_ws_clearClientMessageQueue($connectionId);
echo "Vidage de la queue: " . ($clearResult ? "SUCCESS" : "FAILED") . "\n";

echo "\n10. Vérification après vidage\n";
$queue = frankenphp_ws_getClientMessageQueue($connectionId);
echo "Nombre de messages en queue après vidage: " . count($queue) . "\n";
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur après vidage: " . $counter . " (reste inchangé)\n";

echo "\n11. Envoi de nouveaux messages après vidage\n";
for ($i = 1; $i <= 3; $i++) {
    frankenphp_ws_send($connectionId, "Nouveau message $i", "/test");
    echo "Nouveau message $i envoyé\n";
}

echo "\n12. Vérification finale\n";
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur final: " . $counter . "\n";
$queue = frankenphp_ws_getClientMessageQueue($connectionId);
echo "Messages en queue finale: " . count($queue) . "\n";

echo "\n13. Désactivation de la queue counter\n";
$disableResult = frankenphp_ws_disableQueueCounter($connectionId);
echo "Désactivation: " . ($disableResult ? "SUCCESS" : "FAILED") . "\n";

echo "\n14. Test d'envoi après désactivation (ne sera plus tracé)\n";
frankenphp_ws_send($connectionId, "Message après désactivation", "/test");
$counter = frankenphp_ws_getClientMessageCounter($connectionId);
echo "Compteur après désactivation: " . $counter . " (reste inchangé)\n";

echo "\n=== Test terminé ===\n";
echo "Consultez les logs Caddy pour voir les traces détaillées de la queue counter.\n";
