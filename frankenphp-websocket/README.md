# FrankenPHP WebSocket Extension

Extension WebSocket pour FrankenPHP avec support des tags, stockage d'informations et comptage de clients.

## Documentation des APIs

- **[API Tags](API_TAGS.md)** - Gestion des tags et envoi de messages groupés
- **[API Stored Information](STORED_INFORMATION_API.md)** - Stockage et recherche d'informations par connexion
- **[API Global Information](GLOBAL_INFORMATION_API.md)** - Stockage global avec expiration
- **[API Clients Count](CLIENTS_COUNT_API.md)** - Comptage des clients connectés
- **[API Send All](SEND_ALL_API.md)** - Envoi massif de messages

## Fonctions principales

### Gestion des connexions
- `frankenphp_ws_getClients(?string $route = null): array` - Liste des clients
- `frankenphp_ws_getClientsCount(?string $route = null): int` - Nombre de clients
- `frankenphp_ws_send(string $connectionId, string $data, ?string $route = null): void` - Envoi de message
- `frankenphp_ws_sendAll(string $data, ?string $route = null): int` - Envoi massif

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

