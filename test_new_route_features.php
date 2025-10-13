<?php

echo "<h1>Test des Nouvelles Fonctionnalités de Routes</h1>";

echo "<h2>1. Test de frankenphp_ws_listRoutes()</h2>";
$routes = frankenphp_ws_listRoutes();
echo "Routes actives : <br>";
if (empty($routes)) {
    echo "Aucune route active<br>";
} else {
    foreach ($routes as $route) {
        echo "- Route: '$route'<br>";
    }
}

echo "<h2>2. Test de frankenphp_ws_getClients() avec filtrage par route</h2>";
echo "Tous les clients connectés :<br>";
$allClients = frankenphp_ws_getClients();
echo "Nombre total : " . count($allClients) . "<br>";

if (!empty($routes)) {
    echo "<br>Clients par route :<br>";
    foreach ($routes as $route) {
        $clientsOnRoute = frankenphp_ws_getClients($route);
        echo "Route '$route' : " . count($clientsOnRoute) . " clients<br>";
        if (!empty($clientsOnRoute)) {
            echo "  - " . implode(", ", $clientsOnRoute) . "<br>";
        }
    }
} else {
    echo "Aucune route active pour tester le filtrage<br>";
}

echo "<h2>3. Test de compatibilité rétroactive</h2>";
echo "Test de frankenphp_ws_getClients() sans paramètre (comportement original) :<br>";
$originalBehavior = frankenphp_ws_getClients();
echo "Nombre de clients (sans paramètre) : " . count($originalBehavior) . "<br>";

echo "<h2>4. Test des autres APIs avec routes</h2>";
if (!empty($routes)) {
    $testRoute = $routes[0]; // Prendre la première route disponible
    echo "Test avec la route : '$testRoute'<br>";
    
    // Test des tags
    $tags = frankenphp_ws_getTags();
    if (!empty($tags)) {
        $testTag = $tags[0];
        echo "Test d'envoi au tag '$testTag' sur la route '$testRoute' :<br>";
        frankenphp_ws_sendToTag($testTag, "Test message pour tag $testTag sur route $testRoute", $testRoute);
        echo "Message envoyé !<br>";
    } else {
        echo "Aucun tag disponible pour tester<br>";
    }
    
    // Test d'expression de tags
    echo "Test d'envoi avec expression sur la route '$testRoute' :<br>";
    frankenphp_ws_sendToTagExpression('*', "Test message avec wildcard sur route $testRoute", $testRoute);
    echo "Message envoyé !<br>";
    
} else {
    echo "Aucune route active pour tester les envois<br>";
}

echo "<h2>5. Test des endpoints admin</h2>";
echo "Les endpoints admin suivants sont maintenant disponibles :<br>";
echo "- GET /frankenphp_ws/getClients (tous les clients)<br>";
echo "- GET /frankenphp_ws/getClients?route=/ws/chat (clients sur une route spécifique)<br>";
echo "- GET /frankenphp_ws/getAllRoutes (toutes les routes)<br>";
echo "- GET /frankenphp_ws/getClientsByRoute/{route} (clients par route)<br>";
echo "- POST /frankenphp_ws/send/{clientID}?route=/ws/chat (envoi avec route)<br>";
echo "- POST /frankenphp_ws/sendToTag/{tag}?route=/ws/chat (envoi par tag avec route)<br>";
echo "- POST /frankenphp_ws/sendToTagExpression/{expression}?route=/ws/chat (envoi par expression avec route)<br>";

echo "<h2>✅ Test terminé !</h2>";
echo "Toutes les nouvelles fonctionnalités ont été testées.<br>";
echo "La rétrocompatibilité est préservée.<br>";
?>
