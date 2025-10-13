# FrankenPHP WebSocket - API de Stockage d'Informations

Cette documentation décrit les nouvelles fonctions de stockage d'informations pour les connexions WebSocket dans FrankenPHP.

## Vue d'ensemble

Le système de stockage d'informations permet de stocker, récupérer et gérer des données associées à chaque connexion WebSocket. Ces informations persistent pendant la durée de vie de la connexion et sont automatiquement nettoyées lors de la déconnexion.

## Fonctions disponibles

### `frankenphp_ws_setStoredInformation(string $connectionId, string $key, string $value): void`

Stocke une information pour une connexion WebSocket spécifique.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket
- `$key` (string) : La clé pour identifier l'information
- `$value` (string) : La valeur à stocker

**Exemple :**
```php
// Stocker l'utilisateur connecté
frankenphp_ws_setStoredInformation($connectionId, 'user_id', '12345');

// Stocker des préférences
frankenphp_ws_setStoredInformation($connectionId, 'language', 'fr');

// Stocker des données JSON
$userData = json_encode(['name' => 'John', 'email' => 'john@example.com']);
frankenphp_ws_setStoredInformation($connectionId, 'user_data', $userData);
```

### `frankenphp_ws_getStoredInformation(string $connectionId, string $key): string`

Récupère une information stockée pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket
- `$key` (string) : La clé de l'information à récupérer

**Retour :**
- (string) : La valeur stockée ou une chaîne vide si l'information n'existe pas

**Exemple :**
```php
// Récupérer l'ID utilisateur
$userId = frankenphp_ws_getStoredInformation($connectionId, 'user_id');

// Récupérer la langue préférée
$language = frankenphp_ws_getStoredInformation($connectionId, 'language');

// Récupérer et décoder des données JSON
$userDataJson = frankenphp_ws_getStoredInformation($connectionId, 'user_data');
if (!empty($userDataJson)) {
    $userData = json_decode($userDataJson, true);
}
```

### `frankenphp_ws_deleteStoredInformation(string $connectionId, string $key): void`

Supprime une information spécifique pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket
- `$key` (string) : La clé de l'information à supprimer

**Exemple :**
```php
// Supprimer une information spécifique
frankenphp_ws_deleteStoredInformation($connectionId, 'user_id');
```

### `frankenphp_ws_clearStoredInformation(string $connectionId): void`

Supprime toutes les informations stockées pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket

**Exemple :**
```php
// Nettoyer toutes les informations d'une connexion
frankenphp_ws_clearStoredInformation($connectionId);
```

### `frankenphp_ws_hasStoredInformation(string $connectionId, string $key): bool`

Vérifie si une information existe pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket
- `$key` (string) : La clé de l'information à vérifier

**Retour :**
- (bool) : `true` si l'information existe, `false` sinon

**Exemple :**
```php
// Vérifier si l'utilisateur est authentifié
if (frankenphp_ws_hasStoredInformation($connectionId, 'user_id')) {
    // L'utilisateur est connecté
    $userId = frankenphp_ws_getStoredInformation($connectionId, 'user_id');
} else {
    // L'utilisateur n'est pas connecté
    frankenphp_ws_send($connectionId, json_encode(['error' => 'Not authenticated']));
}
```

### `frankenphp_ws_listStoredInformationKeys(string $connectionId): array`

Liste toutes les clés d'informations stockées pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket

**Retour :**
- (array) : Un tableau contenant toutes les clés d'informations

**Exemple :**
```php
// Lister toutes les clés d'informations
$keys = frankenphp_ws_listStoredInformationKeys($connectionId);

foreach ($keys as $key) {
    $value = frankenphp_ws_getStoredInformation($connectionId, $key);
    echo "Clé: $key, Valeur: $value\n";
}
```

## Cas d'usage pratiques

### 1. Authentification utilisateur

```php
// Lors de la connexion WebSocket
function handleConnection($connectionId) {
    // Vérifier si l'utilisateur est authentifié
    if (frankenphp_ws_hasStoredInformation($connectionId, 'user_id')) {
        $userId = frankenphp_ws_getStoredInformation($connectionId, 'user_id');
        frankenphp_ws_send($connectionId, json_encode(['status' => 'authenticated', 'user_id' => $userId]));
    } else {
        frankenphp_ws_send($connectionId, json_encode(['status' => 'unauthenticated']));
    }
}
```

### 2. Gestion des sessions

```php
// Stocker les informations de session
frankenphp_ws_setStoredInformation($connectionId, 'session_id', session_id());
frankenphp_ws_setStoredInformation($connectionId, 'login_time', date('Y-m-d H:i:s'));

// Récupérer les informations de session
$sessionId = frankenphp_ws_getStoredInformation($connectionId, 'session_id');
$loginTime = frankenphp_ws_getStoredInformation($connectionId, 'login_time');
```

### 3. Préférences utilisateur

```php
// Stocker les préférences
frankenphp_ws_setStoredInformation($connectionId, 'theme', 'dark');
frankenphp_ws_setStoredInformation($connectionId, 'notifications', 'enabled');
frankenphp_ws_setStoredInformation($connectionId, 'language', 'fr');

// Appliquer les préférences
$theme = frankenphp_ws_getStoredInformation($connectionId, 'theme');
$notifications = frankenphp_ws_getStoredInformation($connectionId, 'notifications');
```

### 4. État de l'application

```php
// Stocker l'état actuel
$gameState = json_encode([
    'level' => 5,
    'score' => 1250,
    'lives' => 3
]);
frankenphp_ws_setStoredInformation($connectionId, 'game_state', $gameState);

// Restaurer l'état
$gameStateJson = frankenphp_ws_getStoredInformation($connectionId, 'game_state');
if (!empty($gameStateJson)) {
    $gameState = json_decode($gameStateJson, true);
}
```

## Compatibilité CLI et Web

Toutes les fonctions de stockage d'informations fonctionnent à la fois :
- **En mode CLI** : Les appels sont redirigés vers l'API admin de Caddy via HTTP
- **En mode Web** : Les appels sont traités directement par le serveur WebSocket

## Gestion automatique de la mémoire

- Les informations stockées sont automatiquement supprimées lors de la fermeture d'une connexion WebSocket
- Le système utilise des mutex pour garantir la sécurité des accès concurrents
- Les données sont stockées en mémoire et ne persistent pas entre les redémarrages du serveur

## Limitations

- Les valeurs stockées sont limitées à des chaînes de caractères (string)
- Pour stocker des objets complexes, utilisez `json_encode()` et `json_decode()`
- Les informations ne persistent pas après un redémarrage du serveur
- La taille des données stockées est limitée par la mémoire disponible

## Endpoints API Admin

Les fonctions utilisent également des endpoints API admin pour le mode CLI :

- `POST /frankenphp_ws/setStoredInformation/{clientID}/{key}` : Stocker une information
- `GET /frankenphp_ws/getStoredInformation/{clientID}/{key}` : Récupérer une information
- `DELETE /frankenphp_ws/deleteStoredInformation/{clientID}/{key}` : Supprimer une information
- `DELETE /frankenphp_ws/clearStoredInformation/{clientID}` : Supprimer toutes les informations
- `GET /frankenphp_ws/hasStoredInformation/{clientID}/{key}` : Vérifier l'existence
- `GET /frankenphp_ws/listStoredInformationKeys/{clientID}` : Lister les clés
- `GET /frankenphp_ws/getAllStoredInformation/{clientID}` : Récupérer toutes les informations

Ces endpoints peuvent être utilisés directement pour l'intégration avec d'autres systèmes.
