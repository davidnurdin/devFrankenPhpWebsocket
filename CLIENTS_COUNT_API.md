# API de Comptage des Clients WebSocket

Cette API permet de compter le nombre de clients connectés au serveur WebSocket, avec la possibilité de filtrer par route.

## Fonction PHP

### `frankenphp_ws_getClientsCount(?string $route = null): int`

Compte le nombre de clients connectés au serveur WebSocket.

**Paramètres :**
- `$route` (string, optionnel) : Route spécifique pour filtrer les clients. Si `null` ou vide, compte tous les clients.

**Retour :**
- `int` : Nombre de clients connectés

**Exemples d'utilisation :**

```php
// Compter tous les clients connectés
$totalClients = frankenphp_ws_getClientsCount();
echo "Total clients connectés : $totalClients\n";

// Compter les clients d'une route spécifique
$chatClients = frankenphp_ws_getClientsCount('/chat');
echo "Clients dans /chat : $chatClients\n";

$apiClients = frankenphp_ws_getClientsCount('/api');
echo "Clients dans /api : $apiClients\n";

// Statistiques complètes
$stats = [
    'total' => frankenphp_ws_getClientsCount(),
    'chat' => frankenphp_ws_getClientsCount('/chat'),
    'api' => frankenphp_ws_getClientsCount('/api'),
    'admin' => frankenphp_ws_getClientsCount('/admin')
];

echo "Statistiques : " . json_encode($stats) . "\n";
```

## API Admin (CLI)

### Endpoint

**GET** `/frankenphp_ws/getClientsCount`

### Paramètres

- `route` (query string, optionnel) : Route spécifique pour filtrer les clients

### Réponses

**Sans filtre (tous les clients) :**
```json
{
  "count": 15
}
```

**Avec filtre par route :**
```json
{
  "count": 8,
  "route": "/chat"
}
```

### Exemples d'utilisation

```bash
# Compter tous les clients
curl http://localhost:2019/frankenphp_ws/getClientsCount

# Compter les clients d'une route spécifique
curl "http://localhost:2019/frankenphp_ws/getClientsCount?route=/chat"

# Compter les clients d'une autre route
curl "http://localhost:2019/frankenphp_ws/getClientsCount?route=/api"
```

## Cas d'usage

### 1. Monitoring en temps réel
```php
// Dashboard de monitoring
$dashboard = [
    'total_online' => frankenphp_ws_getClientsCount(),
    'chat_users' => frankenphp_ws_getClientsCount('/chat'),
    'api_consumers' => frankenphp_ws_getClientsCount('/api'),
    'admin_panel' => frankenphp_ws_getClientsCount('/admin')
];

// Envoyer les stats via WebSocket
frankenphp_ws_sendToTag('admin', json_encode([
    'type' => 'stats_update',
    'data' => $dashboard
]));
```

### 2. Limitation de connexions par route
```php
// Vérifier si une route a atteint sa limite
$maxChatUsers = 100;
$currentChatUsers = frankenphp_ws_getClientsCount('/chat');

if ($currentChatUsers >= $maxChatUsers) {
    // Refuser la nouvelle connexion ou rediriger
    frankenphp_ws_send($connectionId, json_encode([
        'error' => 'Chat room is full',
        'current_users' => $currentChatUsers,
        'max_users' => $maxChatUsers
    ]));
}
```

### 3. Statistiques pour les administrateurs
```php
// Fonction pour générer un rapport
function generateConnectionReport() {
    $routes = frankenphp_ws_listRoutes();
    $report = [];
    
    foreach ($routes as $route) {
        $report[$route] = frankenphp_ws_getClientsCount($route);
    }
    
    $report['total'] = frankenphp_ws_getClientsCount();
    
    return $report;
}

// Utilisation
$report = generateConnectionReport();
echo "Rapport de connexions :\n";
foreach ($report as $route => $count) {
    echo "- $route : $count clients\n";
}
```

### 4. Alertes de charge
```php
// Système d'alerte basé sur le nombre de connexions
$totalClients = frankenphp_ws_getClientsCount();
$alertThreshold = 1000;

if ($totalClients > $alertThreshold) {
    // Envoyer une alerte aux administrateurs
    $adminClients = frankenphp_ws_getClientsByTag('admin');
    foreach ($adminClients as $adminId) {
        frankenphp_ws_send($adminId, json_encode([
            'type' => 'alert',
            'message' => "High server load: $totalClients clients connected",
            'threshold' => $alertThreshold
        ]));
    }
}
```

## Compatibilité

- **Mode CLI** : Utilise l'API admin de Caddy via HTTP
- **Mode Worker** : Appel direct aux fonctions Go
- **Thread-safe** : Utilise des mutex pour la sécurité concurrente

## Notes importantes

- Le comptage est effectué en temps réel sur les connexions actives
- Les connexions fermées ne sont pas comptées
- Le paramètre `route` est optionnel et peut être `null`
- La fonction est optimisée pour les performances (lecture seule avec mutex)
- Compatible avec toutes les routes WebSocket configurées

## Limitations

- Ne compte que les connexions WebSocket actives
- Les données ne persistent pas après redémarrage du serveur
- Le comptage est basé sur les connexions en mémoire uniquement
