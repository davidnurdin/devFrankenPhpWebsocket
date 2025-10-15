# FrankenPHP WebSocket Extension

Extension WebSocket pour FrankenPHP avec support des tags, stockage d'informations et comptage de clients.

## Documentation des APIs

- **[API Tags](API_TAGS.md)** - Gestion des tags et envoi de messages groupés
- **[API Stored Information](STORED_INFORMATION_API.md)** - Stockage et recherche d'informations par connexion
- **[API Global Information](GLOBAL_INFORMATION_API.md)** - Stockage global avec expiration
- **[API Clients Count](CLIENTS_COUNT_API.md)** - Comptage des clients connectés
- **[API Send All](SEND_ALL_API.md)** - Envoi massif de messages
- **[API Kill Connection](KILL_CONNECTION_API.md)** - Fermeture forcée de connexions
- **[API Ping Time](PING_TIME_API.md)** - Mesure du temps de ping/pong
- **[API Ping Périodique](PERIODIC_PING_API.md)** - Ping automatique à intervalles réguliers
- **[API Queue Counter](QUEUE_COUNTER_API.md)** - Compteur et queue des messages par client

## Fonctions principales

### Gestion des connexions
- `frankenphp_ws_getClients(?string $route = null): array` - Liste des clients
- `frankenphp_ws_getClientsCount(?string $route = null): int` - Nombre de clients
- `frankenphp_ws_send(string $connectionId, string $data, ?string $route = null): void` - Envoi de message
- `frankenphp_ws_sendAll(string $data, ?string $route = null): int` - Envoi massif
- `frankenphp_ws_killConnection(string $connectionId): bool` - Fermeture forcée
- `frankenphp_ws_getClientPingTime(string $connectionId): int` - Temps de ping
- `frankenphp_ws_enablePing(string $connectionId, int $intervalMs = 0): bool` - Activer ping/pong avec intervalle optionnel
- `frankenphp_ws_disablePing(string $connectionId): bool` - Désactiver ping/pong

### Gestion de la queue counter
- `frankenphp_ws_enableQueueCounter(string $connectionId, int $maxMessages = 100, int $maxTimeSeconds = 3600): bool` - Activer le compteur et la queue des messages
- `frankenphp_ws_disableQueueCounter(string $connectionId): bool` - Désactiver le compteur et la queue des messages
- `frankenphp_ws_getClientMessageCounter(string $connectionId): int` - Obtenir le compteur de messages pour un client
- `frankenphp_ws_getClientMessageQueue(string $connectionId): array` - Obtenir la queue des messages pour un client
- `frankenphp_ws_clearClientMessageQueue(string $connectionId): bool` - Vider la queue des messages pour un client

### Système de tags
- `frankenphp_ws_tagClient(string $connectionId, string $tag): void` - Ajouter un tag
- `frankenphp_ws_untagClient(string $connectionId, string $tag): void` - Retirer un tag
- `frankenphp_ws_getClientsByTag(string $tag): array` - Clients par tag
- `frankenphp_ws_getTagCount(string $tag): int` - Nombre de clients par tag
- `frankenphp_ws_sendToTag(string $tag, string $data, ?string $route = null): void` - Envoi groupé

### Stockage d'informations
- `frankenphp_ws_setStoredInformation(string $connectionId, string $key, string $value): void`
- `frankenphp_ws_getStoredInformation(string $connectionId, string $key): string`
- `frankenphp_ws_searchStoredInformation(string $key, string $op, string $value, ?string $route = null): array`

### Stockage global
- `frankenphp_ws_global_set(string $key, string $value, int $expireSeconds = 0): void`
- `frankenphp_ws_global_get(string $key): string`
- `frankenphp_ws_global_has(string $key): bool`
- `frankenphp_ws_global_delete(string $key): bool`

## Exemple d'utilisation

```php
// Compter les clients
$totalClients = frankenphp_ws_getClientsCount();
$chatClients = frankenphp_ws_getClientsCount('/chat');

// Envoi massif
$sentCount = frankenphp_ws_sendAll("Message à tous les clients");
$chatSentCount = frankenphp_ws_sendAll("Message au chat", "/chat");

// Fermeture forcée
$success = frankenphp_ws_killConnection("client_123");

// Temps de ping
frankenphp_ws_enablePing("client_123"); // Activer le ping/pong et envoyer un ping unique
frankenphp_ws_enablePing("client_123", 5000); // Activer le ping/pong avec ping toutes les 5 secondes
$pingTimeNs = frankenphp_ws_getClientPingTime("client_123");
$pingTimeMs = $pingTimeNs / 1000000; // Convertir en millisecondes
frankenphp_ws_disablePing("client_123"); // Désactiver le ping (arrête aussi le ping périodique)

// Queue counter
frankenphp_ws_enableQueueCounter("client_123", 50, 1800); // Activer la queue (50 messages max, 30 min max)
$counter = frankenphp_ws_getClientMessageCounter("client_123"); // Obtenir le compteur
$queue = frankenphp_ws_getClientMessageQueue("client_123"); // Obtenir la queue des messages
frankenphp_ws_clearClientMessageQueue("client_123"); // Vider la queue
frankenphp_ws_disableQueueCounter("client_123"); // Désactiver la queue

// Gestion des tags
frankenphp_ws_tagClient($connectionId, 'premium');
$premiumCount = frankenphp_ws_getTagCount('premium');
frankenphp_ws_sendToTag('premium', 'Message pour les premium');

// Stockage d'informations
frankenphp_ws_setStoredInformation($connectionId, 'username', 'john_doe');
$users = frankenphp_ws_searchStoredInformation('username', 'eq', 'john_doe');
```

## Test de connexion

```bash
wscat -c "ws://127.0.0.1:5000/ws"
```

