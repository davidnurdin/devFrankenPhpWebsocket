# Changelog - Support des Routes WebSocket

## Version 2.0 - Nouvelles Fonctionnalités de Routes

### ✨ Nouvelles Fonctionnalités

#### 1. Filtrage par Route pour `frankenphp_ws_getClients()`
- **Avant** : `frankenphp_ws_getClients(): array`
- **Maintenant** : `frankenphp_ws_getClients(?string $route = null): array`
- **Fonctionnalité** : Permet de récupérer tous les clients ou uniquement ceux connectés sur une route spécifique

#### 2. Nouvelle API `frankenphp_ws_listRoutes()`
- **Signature** : `frankenphp_ws_listRoutes(): array`
- **Fonctionnalité** : Retourne toutes les routes actives (où au moins un client est connecté)

### 🔧 Fonctionnalités Étendues

#### Paramètre Route Optionnel pour les Fonctions d'Envoi
Toutes les fonctions d'envoi acceptent maintenant un paramètre `route` optionnel :

- `frankenphp_ws_send(connectionId, data, route?)`
- `frankenphp_ws_sendToTag(tag, data, route?)`
- `frankenphp_ws_sendToTagExpression(expression, data, route?)`

### 🌐 Nouveaux Endpoints Admin API

#### Endpoints Modifiés
- `GET /frankenphp_ws/getClients?route=/ws/chat` - Filtrage par route

#### Endpoints Existants (déjà implémentés)
- `GET /frankenphp_ws/getAllRoutes` - Liste toutes les routes
- `GET /frankenphp_ws/getClientsByRoute/{route}` - Clients par route

### 📁 Fichiers Modifiés

#### Code Go
- `frankenphp-websocket/websocket.go`
  - Modifié `frankenphp_ws_getClients()` pour accepter le paramètre route
  - Ajouté `frankenphp_ws_listRoutes()`

- `frankenphp-websocket/caddy.go`
  - Ajouté `WSGetAllRoutes()`
  - Modifié l'endpoint admin `/getClients` pour supporter le filtrage par route

#### Code C
- `frankenphp-websocket/websocket.h`
  - Modifié la signature de `frankenphp_ws_getClients()`
  - Ajouté la signature de `frankenphp_ws_listRoutes()`

- `frankenphp-websocket/websocket.c`
  - Modifié `PHP_FUNCTION(frankenphp_ws_getClients)`
  - Ajouté `PHP_FUNCTION(frankenphp_ws_listRoutes)`

#### Définitions PHP
- `frankenphp-websocket/websocket.stub.php`
  - Modifié `frankenphp_ws_getClients(?string $route = null): array`
  - Ajouté `frankenphp_ws_listRoutes(): array`

- `frankenphp-websocket/websocket_arginfo.h`
  - Mis à jour les signatures des fonctions
  - Ajouté les définitions pour la nouvelle fonction

### 🧪 Tests et Documentation

#### Nouveaux Fichiers de Test
- `test_routes.php` - Test complet des fonctionnalités de routes
- `test_new_route_features.php` - Test spécifique des nouvelles fonctionnalités

#### Documentation
- `ROUTE_SUPPORT.md` - Documentation complète des fonctionnalités de routes
- `CHANGELOG_ROUTES.md` - Ce fichier de changelog

### ✅ Rétrocompatibilité

- **100% rétrocompatible** : Toutes les fonctions existantes continuent de fonctionner sans modification
- **Paramètres optionnels** : Le paramètre `route` est optionnel avec une valeur par défaut `null`
- **Comportement par défaut** : Sans spécifier de route, le comportement reste identique à la version précédente

### 🔄 Migration

Aucune migration n'est nécessaire. Le code existant continue de fonctionner :

```php
// Code existant (continue de fonctionner)
$clients = frankenphp_ws_getClients();

// Nouvelles fonctionnalités (optionnelles)
$chatClients = frankenphp_ws_getClients('/ws/chat');
$routes = frankenphp_ws_listRoutes();
```

### 🚀 Utilisation

#### Récupération des Clients
```php
// Tous les clients (comportement original)
$allClients = frankenphp_ws_getClients();

// Clients sur une route spécifique
$chatClients = frankenphp_ws_getClients('/ws/chat');
```

#### Liste des Routes
```php
// Récupérer toutes les routes actives
$routes = frankenphp_ws_listRoutes();
foreach ($routes as $route) {
    $clients = frankenphp_ws_getClients($route);
    echo "Route '$route' : " . count($clients) . " clients\n";
}
```

#### Envoi avec Filtrage par Route
```php
// Envoi à tous les clients (comportement original)
frankenphp_ws_sendToTag('admin', 'Message pour tous les admins');

// Envoi uniquement aux admins sur la route /ws/chat
frankenphp_ws_sendToTag('admin', 'Message pour les admins du chat', '/ws/chat');
```

### 🎯 Avantages

1. **Flexibilité** : Possibilité de cibler des clients sur des routes spécifiques
2. **Organisation** : Meilleure organisation des connexions par contexte (chat, notifications, etc.)
3. **Performance** : Envoi ciblé réduit le trafic réseau
4. **Sécurité** : Isolation des messages par route
5. **Monitoring** : Visibilité sur les routes actives et leur utilisation
