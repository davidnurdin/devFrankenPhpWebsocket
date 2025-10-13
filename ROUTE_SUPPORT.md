# Support des Routes WebSocket

## Vue d'ensemble

Le support des routes WebSocket permet maintenant de :
- **Stocker automatiquement la route** lors de la connexion d'un client WebSocket
- **Filtrer les envois de messages par route** avec un paramètre optionnel
- **Envoyer des messages à tous les clients** (comportement par défaut) ou **à des clients spécifiques sur une route donnée**

## Nouvelles Fonctionnalités

### 1. Stockage Automatique des Routes

Chaque connexion WebSocket stocke automatiquement la route sur laquelle elle s'est connectée (ex: `/ws/chat`, `/ws/notifications`, etc.).

### 2. Paramètre Route Optionnel

Toutes les fonctions d'envoi de messages et de récupération de clients acceptent maintenant un paramètre `route` optionnel :

#### `frankenphp_ws_getClients(route?)`
- **Avant** : `frankenphp_ws_getClients(): array`
- **Maintenant** : `frankenphp_ws_getClients(?string $route = null): array`

#### `frankenphp_ws_send(connectionId, data, route?)`
- **Avant** : `frankenphp_ws_send(string $connectionId, string $data)`
- **Maintenant** : `frankenphp_ws_send(string $connectionId, string $data, ?string $route = null)`

#### `frankenphp_ws_sendToTag(tag, data, route?)`
- **Avant** : `frankenphp_ws_sendToTag(string $tag, string $data)`
- **Maintenant** : `frankenphp_ws_sendToTag(string $tag, string $data, ?string $route = null)`

#### `frankenphp_ws_sendToTagExpression(expression, data, route?)`
- **Avant** : `frankenphp_ws_sendToTagExpression(string $expression, string $data)`
- **Maintenant** : `frankenphp_ws_sendToTagExpression(string $expression, string $data, ?string $route = null)`

### 3. Nouvelle API pour les Routes

#### `frankenphp_ws_listRoutes(): array`
Retourne un tableau de toutes les routes actives (où au moins un client est connecté).

### 4. Logique de Filtrage

- **Si `route` est `null` ou vide** : Le message est envoyé à tous les clients correspondants (comportement par défaut)
- **Si `route` est spécifiée** : Le message est envoyé uniquement aux clients connectés sur cette route spécifique

## Exemples d'Utilisation

### Exemple 1 : Récupération des clients et routes
```php
// Récupère tous les clients connectés (toutes routes)
$allClients = frankenphp_ws_getClients();

// Récupère uniquement les clients connectés sur /ws/chat
$chatClients = frankenphp_ws_getClients('/ws/chat');

// Récupère toutes les routes actives
$activeRoutes = frankenphp_ws_listRoutes();
```

### Exemple 2 : Envoi à tous les clients
```php
// Envoie à tous les clients connectés (toutes routes)
frankenphp_ws_send($clientId, "Message pour tous");
frankenphp_ws_sendToTag('admin', "Message pour tous les admins");
```

### Exemple 3 : Envoi à une route spécifique
```php
// Envoie uniquement aux clients connectés sur /ws/chat
frankenphp_ws_send($clientId, "Message chat", "/ws/chat");
frankenphp_ws_sendToTag('admin', "Message admin chat", "/ws/chat");
frankenphp_ws_sendToTagExpression('admin&moderator', "Message modérateur chat", "/ws/chat");
```

### Exemple 4 : Dans le worker WebSocket
```php
$handler = static function (array $event): array {
    // L'événement contient maintenant la route
    $route = $event['Route']; // ex: "/ws/chat"
    $connectionId = $event['Connection'];
    
    // Envoyer un message de retour sur la même route
    frankenphp_ws_send($connectionId, "Message reçu sur $route", $route);
    
    return ['ok' => true];
};
```

### Exemple 5 : Gestion complète des routes
```php
// Récupérer toutes les routes actives
$routes = frankenphp_ws_listRoutes();
echo "Routes actives : " . implode(", ", $routes) . "\n";

// Pour chaque route, afficher les clients connectés
foreach ($routes as $route) {
    $clients = frankenphp_ws_getClients($route);
    echo "Route '$route' : " . count($clients) . " clients\n";
    
    // Envoyer un message spécifique à cette route
    frankenphp_ws_sendToTag('admin', "Message pour les admins sur $route", $route);
}
```

## API Admin Endpoints

### Nouveaux Endpoints

#### `GET /frankenphp_ws/getAllRoutes`
Retourne toutes les routes actives :
```json
{
  "routes": ["/ws/chat", "/ws/notifications", "/ws/game"]
}
```

#### `GET /frankenphp_ws/getClientsByRoute/{route}`
Retourne tous les clients connectés sur une route spécifique :
```json
{
  "route": "/ws/chat",
  "clients": ["client1", "client2"],
  "count": 2
}
```

### Endpoints Modifiés

L'endpoint de récupération des clients accepte maintenant un paramètre `route` en query parameter :

#### `GET /frankenphp_ws/getClients?route=/ws/chat`

Tous les endpoints d'envoi acceptent maintenant un paramètre `route` en query parameter :

#### `POST /frankenphp_ws/send/{clientID}?route=/ws/chat`
#### `POST /frankenphp_ws/sendToTag/{tag}?route=/ws/chat`
#### `POST /frankenphp_ws/sendToTagExpression/{expression}?route=/ws/chat`

## Architecture Technique

### Stockage des Routes
- **`connRoutes`** : Map `connectionID -> route` pour stocker la route de chaque connexion
- **`tempRoutes`** : Stockage temporaire par adresse IP pendant l'établissement de la connexion

### Flux de Connexion
1. **Middleware Caddy** : Capture la route depuis `r.URL.Path` et l'ajoute aux headers
2. **CheckOrigin** : Stocke temporairement la route par adresse IP
3. **OnOpen** : Récupère la route temporaire et la stocke définitivement avec la connexion
4. **OnClose** : Nettoie la route de la connexion fermée

### Filtrage des Messages
- **Vérification de route** : Avant l'envoi, vérification que le client est bien sur la route demandée
- **Log de sécurité** : Messages d'avertissement si tentative d'envoi sur une mauvaise route

## Compatibilité

✅ **Rétrocompatible** : Toutes les fonctions existantes continuent de fonctionner sans modification
✅ **Paramètre optionnel** : Le paramètre `route` est optionnel avec une valeur par défaut `null`
✅ **Comportement par défaut** : Envoi à tous les clients si aucune route n'est spécifiée

## Tests

Utilisez le fichier `test_routes.php` pour tester les nouvelles fonctionnalités :

```bash
php test_routes.php
```

Ce fichier démontre :
- Envoi à tous les clients (toutes routes)
- Envoi à une route spécifique
- Combinaison tags + routes
- Combinaison expressions + routes
