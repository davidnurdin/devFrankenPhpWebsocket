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

### 6. Compter les clients par tag
**GET** `/frankenphp_ws/getTagCount/{tag}`

Compte le nombre de clients ayant un tag spécifique.

**Paramètres :**
- `tag` (URL) : Le nom du tag

**Réponse :**
```json
{
  "tag": "room_general",
  "count": 5
}
```

**Exemple :**
```bash
curl http://localhost:2019/frankenphp_ws/getTagCount/room_general
```

### 7. Récupérer tous les tags
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

### 8. Compter les clients par tag
**GET** `/frankenphp_ws/getTagCount/{tag}`

Compte le nombre de clients ayant un tag spécifique.

**Paramètres :**
- `tag` (URL) : Le nom du tag

**Réponse :**
```json
{
  "tag": "room_general",
  "count": 5
}
```

**Exemple :**
```bash
curl http://localhost:2019/frankenphp_ws/getTagCount/room_general
```

### 9. Renommer une connexion
**POST** `/frankenphp_ws/renameConnection/{currentId}/{newId}`

Renomme une connexion WebSocket en changeant son ID tout en préservant toutes les données associées.

**Paramètres :**
- `currentId` (URL) : L'ID actuel de la connexion WebSocket
- `newId` (URL) : Le nouvel ID pour la connexion

**Réponse de succès :**
```json
{
  "success": true,
  "currentId": "old_connection_123",
  "newId": "user_456",
  "message": "Connection renamed successfully"
}
```

**Réponse d'échec :**
```json
{
  "error": "failed to rename connection"
}
```

**Exemple :**
```bash
curl -X POST http://localhost:2019/frankenphp_ws/renameConnection/old_123/new_456
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

### `frankenphp_ws_getTagCount(string $tag): int`
Compte le nombre de clients ayant un tag spécifique.

### `frankenphp_ws_sendToTag(string $tag, string $data): void`
Envoie un message à tous les clients ayant un tag spécifique.

### `frankenphp_ws_renameConnection(string $currentId, string $newId): bool`
Renomme une connexion WebSocket en changeant son ID tout en préservant toutes les données associées.

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

### Comptage de clients par tag
```php
// Compter le nombre de clients ayant un tag spécifique
$roomGeneralCount = frankenphp_ws_getTagCount('room_general');
echo "Nombre de clients dans room_general : $roomGeneralCount\n";

$premiumCount = frankenphp_ws_getTagCount('premium');
echo "Nombre de clients premium : $premiumCount\n";

// Utilisation pratique pour les statistiques
$stats = [
    'total_online' => count(frankenphp_ws_getClients()),
    'room_general' => frankenphp_ws_getTagCount('room_general'),
    'room_private' => frankenphp_ws_getTagCount('room_private'),
    'premium_users' => frankenphp_ws_getTagCount('premium'),
    'admin_users' => frankenphp_ws_getTagCount('admin')
];

echo "Statistiques : " . json_encode($stats) . "\n";
```

### Renommage de connexions
```php
// Renommer une connexion avec l'ID utilisateur
$success = frankenphp_ws_renameConnection('temp_connection_123', 'user_456');

if ($success) {
    // Toutes les données sont préservées (tags, informations stockées, route)
    frankenphp_ws_send('user_456', json_encode(['message' => 'Connection renamed']));
    
    // Les tags existants sont automatiquement transférés
    $premiumClients = frankenphp_ws_getClientsByTag('premium');
    // 'user_456' sera dans la liste si l'ancienne connexion avait le tag 'premium'
}
```

## Notes importantes

- Les tags sont automatiquement supprimés quand une connexion WebSocket se ferme
- Un client peut avoir plusieurs tags
- Les tags sont stockés en mémoire et ne persistent pas après redémarrage du serveur
- Toutes les opérations de tags sont thread-safe grâce aux mutex
