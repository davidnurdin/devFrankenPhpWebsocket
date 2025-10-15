<?php

/**
 * Exemple de worker WebSocket avec gestion des connexions fantômes
 * 
 * Ce worker démontre comment gérer les événements de connexions fantômes
 * dans un contexte réel d'application.
 */

// Configuration
$ghostConnections = []; // Stockage des connexions fantômes
$pendingReconnections = []; // Connexions en attente de reconnexion

echo "Worker WebSocket démarré avec support des connexions fantômes\n";

while (true) {
    $event = frankenphp_ws_get_event();
    
    if (!$event) {
        usleep(10000); // 10ms
        continue;
    }
    
    $type = $event['type'];
    $connectionId = $event['connection'];
    $route = $event['route'] ?? '/unknown';
    $payload = $event['payload'] ?? null;
    
    echo "[" . date('Y-m-d H:i:s') . "] Événement: $type | Connexion: $connectionId | Route: $route\n";
    
    switch ($type) {
        case 'open':
            handleOpen($connectionId, $route);
            break;
            
        case 'message':
            handleMessage($connectionId, $payload, $route);
            break;
            
        case 'ghostConnectionClose':
            handleGhostConnectionClose($connectionId, $payload, $route);
            break;
            
        case 'beforeClose':
            handleBeforeClose($connectionId, $payload, $route);
            break;
            
        case 'close':
            handleClose($connectionId, $payload, $route);
            break;
            
        default:
            echo "  → Événement non géré: $type\n";
    }
}

/**
 * Gestion de l'ouverture de connexion
 */
function handleOpen($connectionId, $route) {
    global $ghostConnections, $pendingReconnections;
    
    echo "  → Nouvelle connexion ouverte\n";
    
    // Vérifier si c'est une reconnexion
    if (isset($pendingReconnections[$connectionId])) {
        echo "  → Reconnexion détectée pour $connectionId\n";
        
        // Libérer l'ancienne connexion fantôme
        $oldConnectionId = $pendingReconnections[$connectionId];
        if (frankenphp_ws_isGhost($oldConnectionId)) {
            echo "  → Libération de l'ancienne connexion fantôme: $oldConnectionId\n";
            frankenphp_ws_releaseGhost($oldConnectionId);
        }
        
        unset($pendingReconnections[$connectionId]);
    }
    
    // Envoyer un message de bienvenue
    frankenphp_ws_send($connectionId, json_encode([
        'type' => 'welcome',
        'message' => 'Connexion établie',
        'connectionId' => $connectionId,
        'timestamp' => time()
    ]));
}

/**
 * Gestion des messages
 */
function handleMessage($connectionId, $payload, $route) {
    echo "  → Message reçu: " . substr($payload, 0, 100) . "\n";
    
    $data = json_decode($payload, true);
    
    if (!$data || !isset($data['type'])) {
        echo "  → Message invalide ignoré\n";
        return;
    }
    
    switch ($data['type']) {
        case 'ping':
            // Répondre au ping
            frankenphp_ws_send($connectionId, json_encode([
                'type' => 'pong',
                'timestamp' => time()
            ]));
            break;
            
        case 'activate_ghost':
            // Activer le mode fantôme
            echo "  → Activation du mode fantôme demandée\n";
            $success = frankenphp_ws_activateGhost($connectionId);
            frankenphp_ws_send($connectionId, json_encode([
                'type' => 'ghost_activated',
                'success' => $success,
                'timestamp' => time()
            ]));
            break;
            
        case 'release_ghost':
            // Libérer le mode fantôme
            echo "  → Libération du mode fantôme demandée\n";
            $success = frankenphp_ws_releaseGhost($connectionId);
            frankenphp_ws_send($connectionId, json_encode([
                'type' => 'ghost_released',
                'success' => $success,
                'timestamp' => time()
            ]));
            break;
            
        case 'check_ghost':
            // Vérifier l'état fantôme
            $isGhost = frankenphp_ws_isGhost($connectionId);
            frankenphp_ws_send($connectionId, json_encode([
                'type' => 'ghost_status',
                'isGhost' => $isGhost,
                'timestamp' => time()
            ]));
            break;
            
        case 'simulate_reconnection':
            // Simuler une reconnexion
            echo "  → Simulation de reconnexion demandée\n";
            $newConnectionId = $connectionId . '-reconnected';
            $pendingReconnections[$newConnectionId] = $connectionId;
            
            // Activer le mode fantôme pour l'ancienne connexion
            frankenphp_ws_activateGhost($connectionId);
            
            frankenphp_ws_send($connectionId, json_encode([
                'type' => 'reconnection_prepared',
                'oldConnectionId' => $connectionId,
                'newConnectionId' => $newConnectionId,
                'timestamp' => time()
            ]));
            break;
            
        default:
            echo "  → Type de message non géré: " . $data['type'] . "\n";
    }
}

/**
 * Gestion de l'événement ghostConnectionClose
 */
function handleGhostConnectionClose($connectionId, $payload, $route) {
    echo "  → Événement ghostConnectionClose reçu\n";
    echo "  → Connexion fantôme libérée: $connectionId\n";
    
    // Ici vous pouvez effectuer des actions spécifiques avant la fermeture
    // Par exemple : sauvegarder l'état, notifier d'autres services, etc.
    
    // Exemple : sauvegarder l'état de la connexion
    $connectionState = [
        'connectionId' => $connectionId,
        'route' => $route,
        'lastActivity' => time(),
        'ghostReleased' => true
    ];
    
    // Sauvegarder dans un fichier ou base de données
    file_put_contents(
        "/tmp/websocket_state_$connectionId.json",
        json_encode($connectionState)
    );
    
    echo "  → État de connexion sauvegardé\n";
}

/**
 * Gestion de l'événement beforeClose
 */
function handleBeforeClose($connectionId, $payload, $route) {
    echo "  → Événement beforeClose reçu\n";
    
    // Nettoyer les ressources spécifiques à cette connexion
    $stateFile = "/tmp/websocket_state_$connectionId.json";
    if (file_exists($stateFile)) {
        unlink($stateFile);
        echo "  → Fichier d'état supprimé\n";
    }
    
    // Notifier les autres connexions si nécessaire
    $clients = frankenphp_ws_getClients();
    foreach ($clients as $clientId) {
        if ($clientId !== $connectionId) {
            frankenphp_ws_send($clientId, json_encode([
                'type' => 'user_disconnected',
                'connectionId' => $connectionId,
                'timestamp' => time()
            ]));
        }
    }
    
    echo "  → Nettoyage terminé\n";
}

/**
 * Gestion de l'événement close
 */
function handleClose($connectionId, $payload, $route) {
    echo "  → Événement close reçu\n";
    echo "  → Connexion fermée définitivement: $connectionId\n";
    
    // Actions finales (logs, statistiques, etc.)
    $logEntry = [
        'timestamp' => date('Y-m-d H:i:s'),
        'event' => 'connection_closed',
        'connectionId' => $connectionId,
        'route' => $route,
        'payload' => $payload
    ];
    
    file_put_contents(
        '/tmp/websocket_closures.log',
        json_encode($logEntry) . "\n",
        FILE_APPEND | LOCK_EX
    );
    
    echo "  → Fermeture enregistrée dans les logs\n";
}

/**
 * Fonction utilitaire pour obtenir les événements
 * (Cette fonction doit être implémentée selon votre système d'événements)
 */
function frankenphp_ws_get_event() {
    // Cette fonction doit être implémentée selon votre système d'événements
    // Elle doit retourner un tableau avec les clés : type, connection, route, payload
    // ou null si aucun événement n'est disponible
    
    // Exemple d'implémentation simplifiée :
    static $eventQueue = [];
    static $eventIndex = 0;
    
    // Simuler quelques événements pour le test
    if (empty($eventQueue)) {
        $eventQueue = [
            [
                'type' => 'open',
                'connection' => 'test-client-1',
                'route' => '/test',
                'payload' => null
            ],
            [
                'type' => 'message',
                'connection' => 'test-client-1',
                'route' => '/test',
                'payload' => json_encode(['type' => 'ping'])
            ]
        ];
    }
    
    if ($eventIndex < count($eventQueue)) {
        return $eventQueue[$eventIndex++];
    }
    
    return null;
}
