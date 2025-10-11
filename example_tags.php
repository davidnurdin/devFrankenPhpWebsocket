<?php

// Exemple d'utilisation du système de tags WebSocket

echo "=== Exemple d'utilisation des tags WebSocket ===\n\n";

// 1. Récupérer la liste des clients connectés
echo "1. Clients connectés :\n";
$clients = frankenphp_ws_getClients();
foreach ($clients as $client) {
    echo "   - $client\n";
}
echo "\n";

// 1.5. Lister tous les tags existants
echo "1.5. Tags existants :\n";
$allTags = frankenphp_ws_getTags();
foreach ($allTags as $tag) {
    echo "   - $tag\n";
}
echo "\n";

// 2. Tagger un client (exemple avec le premier client)
if (!empty($clients)) {
    $firstClient = $clients[0];
    
    echo "2. Tagger le client '$firstClient' avec les tags 'premium' et 'admin' :\n";
    frankenphp_ws_tagClient($firstClient, 'premium');
    frankenphp_ws_tagClient($firstClient, 'admin');
    echo "   ✅ Client tagué\n\n";
    
    // 3. Récupérer les clients avec le tag 'premium'
    echo "3. Clients avec le tag 'premium' :\n";
    $premiumClients = frankenphp_ws_getClientsByTag('premium');
    foreach ($premiumClients as $client) {
        echo "   - $client\n";
    }
    echo "\n";
    

    // 3A. Afficher tags
    echo "3A. Afficher tags : \n";
    $tagsAfter = frankenphp_ws_getTags();
    foreach ($tagsAfter as $tag) {
        echo "   - $tag\n";
    }
    echo "\n";


    // 4. Envoyer un message à tous les clients premium (nouvelle méthode)
    echo "4. Envoyer un message à tous les clients 'premium' (méthode directe) :\n";
    frankenphp_ws_sendToTag('premium', "Message spécial pour les clients premium!");
    echo "   ✅ Message envoyé à tous les clients premium\n\n";
    
    // 5. Retirer le tag 'admin' du client
    echo "5. Retirer le tag 'admin' du client '$firstClient' :\n";
    frankenphp_ws_untagClient($firstClient, 'admin');
    echo "   ✅ Tag 'admin' retiré\n\n";
    
    // 6. Supprimer tous les tags du client
    echo "6. Supprimer tous les tags du client '$firstClient' :\n";
    frankenphp_ws_clearTagClient($firstClient);
    echo "   ✅ Tous les tags supprimés\n\n";
    
    // 7. Lister les tags après suppression
    echo "7. Tags après suppression :\n";
    $tagsAfter = frankenphp_ws_getTags();
    foreach ($tagsAfter as $tag) {
        echo "   - $tag\n";
    }
    echo "\n";
} else {
    echo "Aucun client connecté pour l'exemple.\n";
}

echo "=== Fin de l'exemple ===\n";
