# FrankenPHP WebSocket - API de Gestion des Connexions

Cette documentation décrit les fonctions de gestion des connexions WebSocket dans FrankenPHP, incluant la nouvelle fonction de renommage de connexions.

## Vue d'ensemble

Le système de gestion des connexions permet de renommer les IDs de connexion WebSocket tout en préservant toutes les données associées (tags, informations stockées, routes, etc.).

## Fonctions disponibles

### `frankenphp_ws_renameConnection(string $currentId, string $newId): bool`

Renomme une connexion WebSocket en changeant son ID tout en préservant toutes les données associées.

**Paramètres :**
- `$currentId` (string) : L'ID actuel de la connexion WebSocket
- `$newId` (string) : Le nouvel ID pour la connexion

**Retour :**
- (bool) : `true` si le renommage a réussi, `false` en cas d'échec

**Cas d'échec :**
- L'ID actuel n'existe pas
- Le nouvel ID existe déjà
- Erreur interne du système

**Exemple :**
```php
// Renommer une connexion
$success = frankenphp_ws_renameConnection('old_connection_123', 'user_456');

if ($success) {
    echo "Connexion renommée avec succès\n";
    
    // La connexion peut maintenant être utilisée avec le nouvel ID
    frankenphp_ws_send('user_456', json_encode(['message' => 'Hello with new ID!']));
} else {
    echo "Échec du renommage\n";
}
```

## Cas d'usage pratiques

### 1. Authentification utilisateur

```php
// Lors de l'authentification, renommer la connexion avec l'ID utilisateur
function authenticateUser($connectionId, $userId) {
    $newId = 'user_' . $userId;
    
    if (frankenphp_ws_renameConnection($connectionId, $newId)) {
        // Stocker des informations utilisateur
        frankenphp_ws_setStoredInformation($newId, 'user_id', $userId);
        frankenphp_ws_setStoredInformation($newId, 'authenticated', 'true');
        
        // Tagger l'utilisateur
        frankenphp_ws_tagClient($newId, 'authenticated');
        frankenphp_ws_tagClient($newId, 'user_' . $userId);
        
        frankenphp_ws_send($newId, json_encode([
            'status' => 'authenticated',
            'new_connection_id' => $newId
        ]));
        
        return true;
    }
    
    return false;
}
```

### 2. Gestion des sessions

```php
// Renommer avec l'ID de session
function bindToSession($connectionId, $sessionId) {
    $newId = 'session_' . $sessionId;
    
    if (frankenphp_ws_renameConnection($connectionId, $newId)) {
        frankenphp_ws_tagClient($newId, 'session_' . $sessionId);
        return true;
    }
    
    return false;
}
```

### 3. Migration de connexions

```php
// Migrer une connexion vers un nouveau système d'ID
function migrateConnection($oldId, $newId) {
    // Vérifier que l'ancienne connexion existe
    if (!in_array($oldId, frankenphp_ws_getClients())) {
        return false;
    }
    
    // Vérifier que le nouvel ID n'existe pas
    if (in_array($newId, frankenphp_ws_getClients())) {
        return false;
    }
    
    // Effectuer le renommage
    return frankenphp_ws_renameConnection($oldId, $newId);
}
```

### 4. Gestion des groupes

```php
// Renommer pour inclure l'appartenance à un groupe
function assignToGroup($connectionId, $groupId) {
    $newId = 'group_' . $groupId . '_' . $connectionId;
    
    if (frankenphp_ws_renameConnection($connectionId, $newId)) {
        frankenphp_ws_tagClient($newId, 'group_' . $groupId);
        return $newId;
    }
    
    return false;
}
```

## Préservation des données

Lors du renommage, toutes les données associées à la connexion sont préservées :

### Tags
```php
// Avant le renommage
frankenphp_ws_tagClient('old_id', 'premium');
frankenphp_ws_tagClient('old_id', 'admin');

// Après le renommage
frankenphp_ws_renameConnection('old_id', 'new_id');

// Les tags sont automatiquement transférés
$tags = frankenphp_ws_getClientsByTag('premium');
// 'new_id' sera dans la liste
```

### Informations stockées
```php
// Avant le renommage
frankenphp_ws_setStoredInformation('old_id', 'user_id', '12345');
frankenphp_ws_setStoredInformation('old_id', 'language', 'fr');

// Après le renommage
frankenphp_ws_renameConnection('old_id', 'new_id');

// Les informations sont automatiquement transférées
$userId = frankenphp_ws_getStoredInformation('new_id', 'user_id'); // '12345'
$language = frankenphp_ws_getStoredInformation('new_id', 'language'); // 'fr'
```

### Route
```php
// La route de la connexion est également préservée
// Si la connexion était sur '/chat', elle reste sur '/chat' après le renommage
```

## Compatibilité CLI et Web

La fonction fonctionne dans les deux modes :

### Mode CLI
```php
// En mode CLI, les appels sont redirigés vers l'API admin de Caddy
$success = frankenphp_ws_renameConnection('old_id', 'new_id');
```

### Mode Web
```php
// En mode Web, les appels sont traités directement par le serveur WebSocket
$success = frankenphp_ws_renameConnection('old_id', 'new_id');
```

## Endpoint API Admin

### `POST /frankenphp_ws/renameConnection/{currentId}/{newId}`

Renomme une connexion WebSocket via l'API HTTP.

**Paramètres :**
- `currentId` (URL) : L'ID actuel de la connexion
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

**Exemple d'utilisation :**
```bash
curl -X POST http://localhost:2019/frankenphp_ws/renameConnection/old_123/new_456
```

## Gestion des erreurs

### Vérification avant renommage
```php
function safeRenameConnection($currentId, $newId) {
    // Vérifier que l'ID actuel existe
    $clients = frankenphp_ws_getClients();
    if (!in_array($currentId, $clients)) {
        throw new Exception("Connection ID '$currentId' does not exist");
    }
    
    // Vérifier que le nouvel ID n'existe pas
    if (in_array($newId, $clients)) {
        throw new Exception("New connection ID '$newId' already exists");
    }
    
    // Effectuer le renommage
    if (!frankenphp_ws_renameConnection($currentId, $newId)) {
        throw new Exception("Failed to rename connection");
    }
    
    return true;
}
```

### Gestion des erreurs avec try-catch
```php
try {
    safeRenameConnection('old_id', 'new_id');
    echo "Renommage réussi\n";
} catch (Exception $e) {
    echo "Erreur: " . $e->getMessage() . "\n";
}
```

## Limitations

- Les IDs de connexion doivent être des chaînes de caractères valides
- Le renommage est atomique : soit il réussit complètement, soit il échoue
- Les IDs ne peuvent pas être vides
- Le système vérifie automatiquement l'unicité des nouveaux IDs

## Sécurité

- Toutes les opérations de renommage sont thread-safe grâce aux mutex
- Les vérifications d'existence sont effectuées avant le renommage
- Aucune donnée n'est perdue lors du processus de renommage
- Les connexions WebSocket restent actives pendant le renommage

## Exemple complet

```php
<?php

// Exemple complet d'utilisation de la fonction de renommage
function handleUserLogin($connectionId, $userId, $userRole) {
    // Créer un nouvel ID basé sur l'utilisateur
    $newId = 'user_' . $userId;
    
    // Vérifier que le renommage est possible
    $clients = frankenphp_ws_getClients();
    if (in_array($newId, $clients)) {
        frankenphp_ws_send($connectionId, json_encode([
            'error' => 'User already connected'
        ]));
        return false;
    }
    
    // Effectuer le renommage
    if (frankenphp_ws_renameConnection($connectionId, $newId)) {
        // Stocker les informations utilisateur
        frankenphp_ws_setStoredInformation($newId, 'user_id', $userId);
        frankenphp_ws_setStoredInformation($newId, 'role', $userRole);
        frankenphp_ws_setStoredInformation($newId, 'login_time', date('Y-m-d H:i:s'));
        
        // Tagger l'utilisateur
        frankenphp_ws_tagClient($newId, 'authenticated');
        frankenphp_ws_tagClient($newId, 'role_' . $userRole);
        frankenphp_ws_tagClient($newId, 'user_' . $userId);
        
        // Envoyer confirmation
        frankenphp_ws_send($newId, json_encode([
            'status' => 'authenticated',
            'new_connection_id' => $newId,
            'user_id' => $userId,
            'role' => $userRole
        ]));
        
        // Notifier les autres utilisateurs si nécessaire
        if ($userRole === 'admin') {
            frankenphp_ws_sendToTag('admin', json_encode([
                'message' => 'New admin connected',
                'user_id' => $userId
            ]));
        }
        
        return true;
    }
    
    return false;
}

// Utilisation
$success = handleUserLogin('temp_connection_123', '456', 'premium');
if ($success) {
    echo "Utilisateur authentifié et connexion renommée\n";
} else {
    echo "Échec de l'authentification\n";
}
```
