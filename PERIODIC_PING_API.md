# API Ping Périodique

Cette API permet de gérer le ping périodique automatique pour les connexions WebSocket.

## Fonctionnalités

- **Ping périodique automatique** : Envoie des pings à intervalles réguliers
- **Intervalle configurable** : Spécifiez l'intervalle en millisecondes
- **Gestion automatique** : Le ping périodique est automatiquement démarré/arrêté avec enablePing/disablePing
- **Thread-safe** : Gestion sécurisée des timers en environnement concurrent

## Fonctions

### `frankenphp_ws_enablePing(string $connectionId, int $intervalMs = 0): bool`

Active le ping/pong pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion WebSocket
- `$intervalMs` (int, optionnel) : Intervalle en millisecondes pour le ping périodique (0 = pas de ping périodique)

**Retour :**
- `bool` : `true` si le ping a été activé avec succès, `false` sinon

**Comportement :**
- Si `$intervalMs = 0` (défaut) : Active seulement le ping/pong manuel
- Si `$intervalMs > 0` : Active le ping/pong ET démarre le ping périodique automatique

### `frankenphp_ws_disablePing(string $connectionId): bool`

Désactive le ping/pong et arrête le ping périodique pour une connexion WebSocket.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion WebSocket

**Retour :**
- `bool` : `true` si le ping a été désactivé avec succès, `false` sinon

**Comportement :**
- Désactive le ping/pong
- Arrête le timer de ping périodique s'il était actif
- Nettoie toutes les données de ping associées

## Exemples d'utilisation

### Ping périodique simple

```php
// Activer le ping avec ping automatique toutes les 5 secondes
$success = frankenphp_ws_enablePing("client_123", 5000);

if ($success) {
    echo "Ping périodique activé toutes les 5 secondes\n";
    
    // Attendre et vérifier le temps de ping
    sleep(6);
    $pingTime = frankenphp_ws_getClientPingTime("client_123");
    echo "Temps de ping: " . ($pingTime / 1000000) . " ms\n";
}

// Désactiver le ping (arrête aussi le ping périodique)
frankenphp_ws_disablePing("client_123");
```

### Ping manuel uniquement

```php
// Activer seulement le ping/pong manuel (pas de ping périodique)
$success = frankenphp_ws_enablePing("client_123");

if ($success) {
    echo "Ping manuel activé\n";
}

// Désactiver
frankenphp_ws_disablePing("client_123");
```

### Gestion de plusieurs connexions

```php
$connections = ["client_1", "client_2", "client_3"];

// Activer le ping périodique pour toutes les connexions
foreach ($connections as $connectionId) {
    frankenphp_ws_enablePing($connectionId, 3000); // Ping toutes les 3 secondes
}

// Attendre un peu
sleep(5);

// Vérifier les temps de ping
foreach ($connections as $connectionId) {
    $pingTime = frankenphp_ws_getClientPingTime($connectionId);
    echo "Client $connectionId: " . ($pingTime / 1000000) . " ms\n";
}

// Désactiver pour toutes les connexions
foreach ($connections as $connectionId) {
    frankenphp_ws_disablePing($connectionId);
}
```

## API REST (Admin)

### POST `/frankenphp_ws/enablePing/{clientID}?interval={ms}`

Active le ping pour un client via l'API admin.

**Paramètres URL :**
- `clientID` : ID du client WebSocket
- `interval` (optionnel) : Intervalle en millisecondes

**Exemple :**
```bash
# Ping manuel uniquement
curl -X POST "http://localhost:2019/frankenphp_ws/enablePing/client_123"

# Ping périodique toutes les 5 secondes
curl -X POST "http://localhost:2019/frankenphp_ws/enablePing/client_123?interval=5000"
```

**Réponse :**
```json
{
    "clientID": "client_123",
    "success": true,
    "action": "enablePing",
    "interval": 5000
}
```

### POST `/frankenphp_ws/disablePing/{clientID}`

Désactive le ping pour un client via l'API admin.

**Exemple :**
```bash
curl -X POST "http://localhost:2019/frankenphp_ws/disablePing/client_123"
```

**Réponse :**
```json
{
    "clientID": "client_123",
    "success": true,
    "action": "disablePing"
}
```

## Notes techniques

### Gestion des timers

- Chaque connexion a son propre timer de ping périodique
- Les timers sont automatiquement nettoyés lors de la désactivation
- En cas d'erreur de ping, le timer est automatiquement arrêté

### Thread safety

- Toutes les opérations sont protégées par des mutex
- Les timers sont gérés de manière thread-safe
- Pas de fuites mémoire lors de la fermeture des connexions

### Performance

- Les timers utilisent `time.AfterFunc` de Go pour une gestion efficace
- Pas d'impact sur les performances des autres connexions
- Nettoyage automatique des ressources

### Valeurs recommandées

- **Intervalle minimum** : 1000ms (1 seconde)
- **Intervalle recommandé** : 3000-10000ms (3-10 secondes)
- **Intervalle maximum** : 60000ms (1 minute)

## Dépannage

### Le ping périodique ne fonctionne pas

1. Vérifiez que la connexion existe
2. Vérifiez que l'intervalle est > 0
3. Consultez les logs pour les erreurs

### Fuites mémoire

- Les timers sont automatiquement nettoyés
- `disablePing` nettoie toutes les ressources
- Pas d'action manuelle requise

### Performance dégradée

- Réduisez la fréquence des pings
- Vérifiez le nombre de connexions actives
- Surveillez l'utilisation CPU/mémoire
