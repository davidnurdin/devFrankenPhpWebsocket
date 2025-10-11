# API Admin WebSocket Tags

Ce document décrit les endpoints API admin pour gérer les tags des connexions WebSocket.

## Endpoints disponibles

### 1. Tagger un client
**POST** `/frankenphp_ws/tag/{clientID}`

Ajoute un tag à une connexion WebSocket.

**Paramètres :**
- `clientID` (URL) : ID de la connexion WebSocket
- Body : Le nom du tag (content-type: text/plain)

**Exemple :**
```bash
curl -X POST http://localhost:2019/frankenphp_ws/tag/abc123 \
  -H "Content-Type: text/plain" \
  -d "premium"
```

### 2. Retirer un tag d'un client
**DELETE** `/frankenphp_ws/untag/{clientID}/{tag}`

Retire un tag spécifique d'une connexion WebSocket.

**Paramètres :**
- `clientID` (URL) : ID de la connexion WebSocket
- `tag` (URL) : Le nom du tag à retirer

**Exemple :**
```bash
curl -X DELETE http://localhost:2019/frankenphp_ws/untag/abc123/premium
```

### 3. Supprimer tous les tags d'un client
**DELETE** `/frankenphp_ws/clearTags/{clientID}`

Supprime tous les tags d'une connexion WebSocket.

**Paramètres :**
- `clientID` (URL) : ID de la connexion WebSocket

**Exemple :**
```bash
curl -X DELETE http://localhost:2019/frankenphp_ws/clearTags/abc123
```

### 4. Récupérer les tags d'un client
**GET** `/frankenphp_ws/getTags/{clientID}`

Récupère tous les tags d'une connexion WebSocket.

**Paramètres :**
- `clientID` (URL) : ID de la connexion WebSocket

**Réponse :**
```json
{
  "clientID": "abc123",
  "tags": ["premium", "admin"]
}
```

**Exemple :**
```bash
curl http://localhost:2019/frankenphp_ws/getTags/abc123
```

### 5. Récupérer les clients par tag
**GET** `/frankenphp_ws/getClientsByTag/{tag}`

Récupère tous les clients ayant un tag spécifique.

**Paramètres :**
- `tag` (URL) : Le nom du tag

**Réponse :**
```json
{
  "tag": "premium",
  "clients": ["abc123", "def456"]
}
```

**Exemple :**
```bash
curl http://localhost:2019/frankenphp_ws/getClientsByTag/premium
```

### 6. Récupérer tous les tags
**GET** `/frankenphp_ws/getAllTags`

Récupère tous les tags existants.

**Réponse :**
```json
{
  "tags": ["premium", "admin", "vip"]
}
```

**Exemple :**
```bash
curl http://localhost:2019/frankenphp_ws/getAllTags
```

### 7. Envoyer un message à tous les clients d'un tag
**POST** `/frankenphp_ws/sendToTag/{tag}`

Envoie un message à tous les clients ayant un tag spécifique.

**Paramètres :**
- `tag` (URL) : Le nom du tag
- Body : Le message à envoyer (content-type: application/octet-stream)

**Réponse :**
```json
{
  "tag": "premium",
  "sentCount": 3
}
```

**Exemple :**
```bash
curl -X POST http://localhost:2019/frankenphp_ws/sendToTag/premium \
  -H "Content-Type: application/octet-stream" \
  -d "Message pour les clients premium"
```

## Fonctions PHP disponibles

### `frankenphp_ws_tagClient(string $connectionId, string $tag): void`
Ajoute un tag à une connexion WebSocket.

### `frankenphp_ws_untagClient(string $connectionId, string $tag): void`
Retire un tag spécifique d'une connexion WebSocket.

### `frankenphp_ws_clearTagClient(string $connectionId): void`
Supprime tous les tags d'une connexion WebSocket.

### `frankenphp_ws_getTags(): array`
Retourne tous les tags existants.

### `frankenphp_ws_getClientsByTag(string $tag): array`
Retourne tous les clients ayant un tag spécifique.

### `frankenphp_ws_sendToTag(string $tag, string $data): void`
Envoie un message à tous les clients ayant un tag spécifique.

## Exemples d'utilisation

### Grouper les clients par rôle
```php
// Tagger les clients selon leur rôle
frankenphp_ws_tagClient($connectionId, 'admin');
frankenphp_ws_tagClient($connectionId, 'premium');

// Envoyer un message à tous les admins (nouvelle méthode directe)
frankenphp_ws_sendToTag('admin', "Message pour les admins");

// Récupérer tous les clients admin
$adminClients = frankenphp_ws_getClientsByTag('admin');
echo "Clients admin : " . implode(', ', $adminClients) . "\n";

// Lister tous les tags existants
$allTags = frankenphp_ws_getTags();
echo "Tags disponibles : " . implode(', ', $allTags) . "\n";
```

### Gestion des sessions
```php
// Tagger un client avec son ID de session
frankenphp_ws_tagClient($connectionId, 'session_' . $sessionId);

// Plus tard, retirer le tag de session
frankenphp_ws_untagClient($connectionId, 'session_' . $sessionId);
```

## Notes importantes

- Les tags sont automatiquement supprimés quand une connexion WebSocket se ferme
- Un client peut avoir plusieurs tags
- Les tags sont stockés en mémoire et ne persistent pas après redémarrage du serveur
- Toutes les opérations de tags sont thread-safe grâce aux mutex
