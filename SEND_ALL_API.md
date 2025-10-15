# API d'Envoi Massif WebSocket

Cette API permet d'envoyer un message à tous les clients connectés au serveur WebSocket, avec la possibilité de filtrer par route.

## Fonction PHP

### `frankenphp_ws_sendAll(string $data, ?string $route = null): int`

Envoie un message à tous les clients connectés au serveur WebSocket.

**Paramètres :**
- `$data` (string) : Le message à envoyer
- `$route` (string, optionnel) : Route spécifique pour filtrer les clients. Si `null` ou vide, envoie à tous les clients.

**Retour :**
- `int` : Nombre de clients ayant reçu le message avec succès

**Exemples d'utilisation :**

```php
// Envoyer un message à tous les clients
$sentCount = frankenphp_ws_sendAll("Message global à tous les clients");
echo "Message envoyé à $sentCount clients\n";

// Envoyer un message à tous les clients d'une route spécifique
$sentCount = frankenphp_ws_sendAll("Message pour le chat", "/chat");
echo "Message envoyé à $sentCount clients du chat\n";

// Envoyer des notifications système
$notification = json_encode([
    'type' => 'system_notification',
    'message' => 'Le serveur va redémarrer dans 5 minutes',
    'timestamp' => time()
]);

$totalSent = frankenphp_ws_sendAll($notification);
echo "Notification envoyée à $totalSent clients\n";

// Envoyer des messages par route
$routes = ['/chat', '/api', '/admin'];
foreach ($routes as $route) {
    $message = "Message spécifique pour la route $route";
    $sentCount = frankenphp_ws_sendAll($message, $route);
    echo "Route $route : $sentCount clients\n";
}
```

## API Admin (CLI)

### Endpoint

**POST** `/frankenphp_ws/sendAll`

### Paramètres

- `route` (query string, optionnel) : Route spécifique pour filtrer les clients
- Body : Le message à envoyer (content-type: application/octet-stream)

### Réponses

**Sans filtre (tous les clients) :**
```json
{
  "sentCount": 15
}
```

**Avec filtre par route :**
```json
{
  "sentCount": 8,
  "route": "/chat"
}
```

### Exemples d'utilisation

```bash
# Envoyer un message à tous les clients
curl -X POST http://localhost:2019/frankenphp_ws/sendAll \
  -H "Content-Type: application/octet-stream" \
  -d "Message global à tous les clients"

# Envoyer un message à une route spécifique
curl -X POST "http://localhost:2019/frankenphp_ws/sendAll?route=/chat" \
  -H "Content-Type: application/octet-stream" \
  -d "Message pour le chat"

# Envoyer un JSON
curl -X POST http://localhost:2019/frankenphp_ws/sendAll \
  -H "Content-Type: application/octet-stream" \
  -d '{"type":"notification","message":"Maintenance programmée"}'
```

## Cas d'usage

### 1. Notifications système
```php
// Fonction pour envoyer des notifications système
function sendSystemNotification($message, $route = null) {
    $notification = json_encode([
        'type' => 'system',
        'message' => $message,
        'timestamp' => time(),
        'id' => uniqid()
    ]);
    
    return frankenphp_ws_sendAll($notification, $route);
}

// Utilisation
$sentCount = sendSystemNotification("Maintenance programmée à 2h du matin");
echo "Notification envoyée à $sentCount clients\n";

// Notification spécifique à une route
$sentCount = sendSystemNotification("Chat en maintenance", "/chat");
```

### 2. Annonces et publicités
```php
// Système d'annonces
function sendAnnouncement($title, $content, $targetRoute = null) {
    $announcement = json_encode([
        'type' => 'announcement',
        'title' => $title,
        'content' => $content,
        'timestamp' => time()
    ]);
    
    return frankenphp_ws_sendAll($announcement, $targetRoute);
}

// Envoyer une annonce à tous les clients
$sentCount = sendAnnouncement(
    "Nouvelle fonctionnalité disponible !",
    "Découvrez notre nouveau système de chat en temps réel."
);

// Annonce spécifique aux utilisateurs premium
$sentCount = sendAnnouncement(
    "Fonctionnalité premium",
    "Accédez à des fonctionnalités avancées.",
    "/premium"
);
```

### 3. Mise à jour de statut
```php
// Fonction pour mettre à jour le statut du serveur
function updateServerStatus($status, $message) {
    $statusUpdate = json_encode([
        'type' => 'server_status',
        'status' => $status, // 'online', 'maintenance', 'offline'
        'message' => $message,
        'timestamp' => time()
    ]);
    
    return frankenphp_ws_sendAll($statusUpdate);
}

// Utilisation
$sentCount = updateServerStatus('maintenance', 'Le serveur sera en maintenance pendant 30 minutes');
echo "Mise à jour de statut envoyée à $sentCount clients\n";
```

### 4. Messages de bienvenue
```php
// Envoyer un message de bienvenue à tous les nouveaux utilisateurs
function sendWelcomeMessage() {
    $welcome = json_encode([
        'type' => 'welcome',
        'message' => 'Bienvenue sur notre plateforme !',
        'features' => [
            'Chat en temps réel',
            'Notifications push',
            'Support 24/7'
        ]
    ]);
    
    return frankenphp_ws_sendAll($welcome);
}

// Message de bienvenue spécifique au chat
function sendChatWelcome() {
    $chatWelcome = json_encode([
        'type' => 'chat_welcome',
        'message' => 'Bienvenue dans le chat ! N\'hésitez pas à poser vos questions.',
        'rules' => [
            'Respectez les autres utilisateurs',
            'Pas de spam',
            'Utilisez un langage approprié'
        ]
    ]);
    
    return frankenphp_ws_sendAll($chatWelcome, '/chat');
}
```

### 5. Alertes et urgences
```php
// Système d'alertes d'urgence
function sendEmergencyAlert($severity, $message, $affectedRoutes = null) {
    $alert = json_encode([
        'type' => 'emergency_alert',
        'severity' => $severity, // 'low', 'medium', 'high', 'critical'
        'message' => $message,
        'timestamp' => time(),
        'requires_acknowledgment' => $severity === 'critical'
    ]);
    
    if ($affectedRoutes) {
        $totalSent = 0;
        foreach ($affectedRoutes as $route) {
            $sentCount = frankenphp_ws_sendAll($alert, $route);
            $totalSent += $sentCount;
        }
        return $totalSent;
    }
    
    return frankenphp_ws_sendAll($alert);
}

// Utilisation
$sentCount = sendEmergencyAlert(
    'high',
    'Problème de sécurité détecté. Veuillez vous déconnecter et vous reconnecter.',
    ['/chat', '/api']
);
```

### 6. Statistiques et monitoring
```php
// Envoyer des statistiques en temps réel
function broadcastServerStats() {
    $stats = [
        'type' => 'server_stats',
        'total_clients' => frankenphp_ws_getClientsCount(),
        'chat_clients' => frankenphp_ws_getClientsCount('/chat'),
        'api_clients' => frankenphp_ws_getClientsCount('/api'),
        'timestamp' => time()
    ];
    
    return frankenphp_ws_sendAll(json_encode($stats));
}

// Envoyer des stats aux administrateurs
function sendAdminStats() {
    $adminStats = [
        'type' => 'admin_stats',
        'total_clients' => frankenphp_ws_getClientsCount(),
        'premium_users' => frankenphp_ws_getTagCount('premium'),
        'admin_users' => frankenphp_ws_getTagCount('admin'),
        'timestamp' => time()
    ];
    
    return frankenphp_ws_sendAll(json_encode($adminStats), '/admin');
}
```

## Compatibilité

- **Mode CLI** : Utilise l'API admin de Caddy via HTTP POST
- **Mode Worker** : Appel direct aux fonctions Go
- **Thread-safe** : Utilise des mutex pour la sécurité concurrente

## Notes importantes

- Le message est envoyé à tous les clients connectés (ou filtrés par route)
- La fonction retourne le nombre de clients ayant reçu le message avec succès
- Les connexions fermées ou en erreur ne sont pas comptées dans le retour
- Le paramètre `route` est optionnel et peut être `null`
- Compatible avec tous les types de données (texte, JSON, binaire)

## Limitations

- Ne peut pas envoyer à des clients spécifiques (utilisez `frankenphp_ws_send` pour cela)
- Les messages ne persistent pas si le client n'est pas connecté
- La taille des messages est limitée par la mémoire disponible
- Les données ne persistent pas après redémarrage du serveur

## Différences avec les autres fonctions d'envoi

| Fonction | Cible | Retour |
|----------|-------|--------|
| `frankenphp_ws_send` | Client spécifique | void |
| `frankenphp_ws_sendToTag` | Clients avec un tag | int (nombre envoyé) |
| `frankenphp_ws_sendAll` | Tous les clients (ou par route) | int (nombre envoyé) |
