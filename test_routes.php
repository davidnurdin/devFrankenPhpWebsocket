<?php

echo "Test des routes WebSocket<br>";

echo "Get all WS clients ... <br>";
$clients = frankenphp_ws_getClients();
echo "Nombre de clients connectés : " . count($clients) . "<br>";

echo "<br>Liste des routes actives :<br>";
$routes = frankenphp_ws_listRoutes();
echo "Routes disponibles : " . implode(", ", $routes) . "<br>";

echo "<br>Test du filtrage par route :<br>";
foreach ($routes as $route) {
    $clientsOnRoute = frankenphp_ws_getClients($route);
    echo "Clients sur la route '$route' : " . count($clientsOnRoute) . "<br>";
}

// Envoyer un message à tous les clients (toutes routes)
echo "<br>Envoi à tous les clients (toutes routes) :<br>";
foreach ($clients as $client) {
    echo "Send HELLO to " . $client . "<br>";
    frankenphp_ws_send($client, "HELLO ALL ROUTES " . rand(1,1000));
}

// Envoyer un message à tous les clients sur une route spécifique
echo "<br>Envoi à tous les clients sur la route /ws/test :<br>";
foreach ($clients as $client) {
    echo "Send HELLO to " . $client . " on route /ws/test<br>";
    frankenphp_ws_send($client, "HELLO ROUTE /ws/test " . rand(1,1000), "/ws/test");
}

// Test avec les tags
echo "<br>Test avec les tags :<br>";
$tags = frankenphp_ws_getTags();
echo "Tags disponibles : " . implode(", ", $tags) . "<br>";

// Envoyer à un tag spécifique (toutes routes)
echo "<br>Envoi au tag group_2 (toutes routes) :<br>";
frankenphp_ws_sendToTag('group_2', 'group 2 TAG - all routes');

// Envoyer à un tag spécifique sur une route spécifique
echo "<br>Envoi au tag group_2 sur la route /ws/test :<br>";
frankenphp_ws_sendToTag('group_2', 'group 2 TAG - route /ws/test', '/ws/test');

// Test avec les expressions de tags
echo "<br>Test avec les expressions de tags :<br>";

// Envoyer à une expression (toutes routes)
echo "<br>Envoi à l'expression group_* (toutes routes) :<br>";
frankenphp_ws_sendToTagExpression('group_*', 'Expression group_* - all routes');

// Envoyer à une expression sur une route spécifique
echo "<br>Envoi à l'expression group_* sur la route /ws/test :<br>";
frankenphp_ws_sendToTagExpression('group_*', 'Expression group_* - route /ws/test', '/ws/test');

echo "<br>Test terminé !<br>";
?>
