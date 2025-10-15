# API de Mesure du Temps de Ping WebSocket

Cette API permet de mesurer le temps de réponse ping/pong des connexions WebSocket, en s'inspirant de la bibliothèque [gws](https://github.com/lxzan/gws/).

## Fonctions PHP

### `frankenphp_ws_enablePing(string $connectionId): bool`

Active le ping/pong pour une connexion WebSocket spécifique.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket

**Retour :**
- `bool` : `true` si le ping a été activé avec succès

### `frankenphp_ws_disablePing(string $connectionId): bool`

Désactive le ping/pong pour une connexion WebSocket spécifique.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket

**Retour :**
- `bool` : `true` si le ping a été désactivé avec succès

### `frankenphp_ws_getClientPingTime(string $connectionId): int`

Récupère le temps de ping (en nanosecondes) d'une connexion WebSocket spécifique.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket

**Retour :**
- `int` : Temps de ping en nanosecondes (0 si pas de données disponibles)

**Exemples d'utilisation :**

```php
// Activer le ping pour un client
$success = frankenphp_ws_enablePing("client_123");
if ($success) {
    echo "Ping activé pour client_123\n";
}

// Désactiver le ping pour un client
$success = frankenphp_ws_disablePing("client_123");
if ($success) {
    echo "Ping désactivé pour client_123\n";
}

// Récupérer le temps de ping d'un client (nécessite que le ping soit activé)
$pingTimeNs = frankenphp_ws_getClientPingTime("client_123");
$pingTimeMs = $pingTimeNs / 1000000; // Convertir en millisecondes
echo "Temps de ping : {$pingTimeMs}ms\n";

// Fonction utilitaire pour convertir en millisecondes
function getPingTimeMs($connectionId) {
    $pingTimeNs = frankenphp_ws_getClientPingTime($connectionId);
    return $pingTimeNs / 1000000; // Convertir en millisecondes
}

// Vérifier la qualité de connexion
function checkConnectionQuality($connectionId) {
    $pingTimeMs = getPingTimeMs($connectionId);
    
    if ($pingTimeMs == 0) {
        return "Pas de données de ping";
    } elseif ($pingTimeMs < 50) {
        return "Excellent ($pingTimeMs ms)";
    } elseif ($pingTimeMs < 100) {
        return "Bon ($pingTimeMs ms)";
    } elseif ($pingTimeMs < 200) {
        return "Moyen ($pingTimeMs ms)";
    } else {
        return "Lent ($pingTimeMs ms)";
    }
}

$quality = checkConnectionQuality("client_123");
echo "Qualité de connexion : $quality\n";
```

## API Admin (CLI)

### Endpoints

**POST** `/frankenphp_ws/enablePing/{clientID}`

Active le ping/pong pour un client spécifique.

**POST** `/frankenphp_ws/disablePing/{clientID}`

Désactive le ping/pong pour un client spécifique.

**GET** `/frankenphp_ws/getClientPingTime/{clientID}`

### Paramètres

- `clientID` (URL) : L'ID de la connexion WebSocket

### Réponses

**Succès :**
```json
{
  "clientID": "client_123",
  "pingTime": 25000000,
  "pingTimeMs": 25.0
}
```

**Pas de données :**
```json
{
  "clientID": "client_123",
  "pingTime": 0,
  "pingTimeMs": 0.0
}
```

### Exemples d'utilisation

```bash
# Activer le ping pour un client
curl -X POST http://localhost:2019/frankenphp_ws/enablePing/client_123

# Désactiver le ping pour un client
curl -X POST http://localhost:2019/frankenphp_ws/disablePing/client_123

# Récupérer le temps de ping d'un client
curl http://localhost:2019/frankenphp_ws/getClientPingTime/client_123

# Récupérer le temps de ping avec un ID complexe
curl "http://localhost:2019/frankenphp_ws/getClientPingTime/user%3A456%3Asession%3A789"
```

## Fonctionnement du système Ping/Pong

### Handlers implémentés

Le système utilise les handlers `OnPing` et `OnPong` de la bibliothèque [gws](https://github.com/lxzan/gws/) :

1. **OnPing** : Déclenché quand le serveur reçoit un ping du client
   - Vérifie si le ping est activé pour ce client
   - Si activé : stocke le timestamp du ping et envoie un pong en réponse
   - Si désactivé : ignore le ping

2. **OnPong** : Déclenché quand le serveur reçoit un pong du client
   - Vérifie si le ping est activé pour ce client
   - Si activé : calcule le temps écoulé depuis le ping correspondant et stocke le temps de réponse
   - Si désactivé : ignore le pong

### Contrôle d'activation

**Par défaut, le ping/pong est désactivé** pour tous les clients. Il faut l'activer explicitement avec `frankenphp_ws_enablePing()`.

### Mesure précise

Le système mesure le temps réel entre l'envoi du ping et la réception du pong :

```go
// Dans OnPing
clientPingTimestamps[connectionID] = time.Now()

// Dans OnPong
pingTime := time.Since(pingTimestamp)
clientPingTimes[connectionID] = pingTime
```

## Cas d'usage

### 1. Monitoring de la qualité de connexion
```php
// Fonction pour surveiller la qualité des connexions
function monitorConnectionQuality() {
    $allClients = frankenphp_ws_getClients();
    $report = [];
    
    foreach ($allClients as $connectionId) {
        $pingTimeMs = getPingTimeMs($connectionId);
        
        if ($pingTimeMs > 0) {
            $report[] = [
                'connectionId' => $connectionId,
                'pingTime' => $pingTimeMs,
                'quality' => $pingTimeMs < 100 ? 'good' : 'poor'
            ];
        }
    }
    
    return $report;
}

// Surveillance en temps réel
$qualityReport = monitorConnectionQuality();
foreach ($qualityReport as $client) {
    if ($client['quality'] === 'poor') {
        echo "Client {$client['connectionId']} a une connexion lente: {$client['pingTime']}ms\n";
    }
}
```

### 2. Détection des connexions lentes
```php
// Fonction pour détecter les connexions lentes
function detectSlowConnections($thresholdMs = 200) {
    $allClients = frankenphp_ws_getClients();
    $slowConnections = [];
    
    foreach ($allClients as $connectionId) {
        $pingTimeMs = getPingTimeMs($connectionId);
        
        if ($pingTimeMs > $thresholdMs) {
            $slowConnections[] = [
                'connectionId' => $connectionId,
                'pingTime' => $pingTimeMs
            ];
        }
    }
    
    return $slowConnections;
}

// Détecter les connexions lentes
$slowConnections = detectSlowConnections(200);
if (!empty($slowConnections)) {
    echo "Connexions lentes détectées :\n";
    foreach ($slowConnections as $conn) {
        echo "- {$conn['connectionId']}: {$conn['pingTime']}ms\n";
    }
}
```

### 3. Statistiques de performance
```php
// Fonction pour générer des statistiques de ping
function generatePingStats() {
    $allClients = frankenphp_ws_getClients();
    $pingTimes = [];
    
    foreach ($allClients as $connectionId) {
        $pingTimeMs = getPingTimeMs($connectionId);
        if ($pingTimeMs > 0) {
            $pingTimes[] = $pingTimeMs;
        }
    }
    
    if (empty($pingTimes)) {
        return null;
    }
    
    sort($pingTimes);
    $count = count($pingTimes);
    
    return [
        'count' => $count,
        'min' => $pingTimes[0],
        'max' => $pingTimes[$count - 1],
        'avg' => array_sum($pingTimes) / $count,
        'median' => $pingTimes[intval($count / 2)]
    ];
}

// Générer et afficher les statistiques
$stats = generatePingStats();
if ($stats) {
    echo "Statistiques de ping :\n";
    echo "- Nombre de clients : {$stats['count']}\n";
    echo "- Ping minimum : {$stats['min']}ms\n";
    echo "- Ping maximum : {$stats['max']}ms\n";
    echo "- Ping moyen : " . round($stats['avg'], 2) . "ms\n";
    echo "- Ping médian : {$stats['median']}ms\n";
}
```

### 4. Alertes de performance
```php
// Système d'alerte pour les connexions lentes
function checkPerformanceAlerts() {
    $allClients = frankenphp_ws_getClients();
    $alerts = [];
    
    foreach ($allClients as $connectionId) {
        $pingTimeMs = getPingTimeMs($connectionId);
        
        if ($pingTimeMs > 500) {
            $alerts[] = [
                'type' => 'critical',
                'connectionId' => $connectionId,
                'pingTime' => $pingTimeMs,
                'message' => "Connexion très lente détectée"
            ];
        } elseif ($pingTimeMs > 200) {
            $alerts[] = [
                'type' => 'warning',
                'connectionId' => $connectionId,
                'pingTime' => $pingTimeMs,
                'message' => "Connexion lente détectée"
            ];
        }
    }
    
    return $alerts;
}

// Vérifier les alertes
$alerts = checkPerformanceAlerts();
foreach ($alerts as $alert) {
    echo "[{$alert['type']}] {$alert['message']} - Client: {$alert['connectionId']} ({$alert['pingTime']}ms)\n";
}
```

### 5. Optimisation des routes
```php
// Fonction pour analyser la performance par route
function analyzeRoutePerformance() {
    $routes = ['/chat', '/api', '/admin']; // Routes à analyser
    $routeStats = [];
    
    foreach ($routes as $route) {
        $clients = frankenphp_ws_getClients($route);
        $pingTimes = [];
        
        foreach ($clients as $connectionId) {
            $pingTimeMs = getPingTimeMs($connectionId);
            if ($pingTimeMs > 0) {
                $pingTimes[] = $pingTimeMs;
            }
        }
        
        if (!empty($pingTimes)) {
            $routeStats[$route] = [
                'clientCount' => count($clients),
                'avgPing' => array_sum($pingTimes) / count($pingTimes),
                'maxPing' => max($pingTimes),
                'minPing' => min($pingTimes)
            ];
        }
    }
    
    return $routeStats;
}

// Analyser les performances par route
$routeStats = analyzeRoutePerformance();
foreach ($routeStats as $route => $stats) {
    echo "Route $route :\n";
    echo "  - Clients : {$stats['clientCount']}\n";
    echo "  - Ping moyen : " . round($stats['avgPing'], 2) . "ms\n";
    echo "  - Ping min/max : {$stats['minPing']}ms / {$stats['maxPing']}ms\n";
}
```

## Compatibilité

- **Mode CLI** : Utilise l'API admin de Caddy via HTTP GET
- **Mode Worker** : Appel direct aux fonctions Go
- **Thread-safe** : Utilise des mutex pour la sécurité concurrente

## Notes importantes

- Le temps de ping est mesuré en **nanosecondes** pour la précision maximale
- Les données de ping sont **automatiquement nettoyées** lors de la fermeture de connexion
- Le système utilise les **handlers natifs** de la bibliothèque [gws](https://github.com/lxzan/gws/)
- Les mesures sont **précises** car basées sur les timestamps système
- Retourne `0` si aucune donnée de ping n'est disponible

## Limitations

- Nécessite que le client envoie des pings pour avoir des données
- Les données ne persistent pas après redémarrage du serveur
- Ne peut pas mesurer le ping si le client ne répond pas aux pings
- Les timestamps sont stockés en mémoire uniquement

## Intégration avec gws

Cette implémentation suit les bonnes pratiques de la bibliothèque [gws](https://github.com/lxzan/gws/) :

- Utilise les handlers `OnPing` et `OnPong` standard
- Respecte le protocole WebSocket pour les frames ping/pong
- Gère automatiquement l'envoi des réponses pong
- Compatible avec les fonctionnalités de compression et de parallélisation de gws
