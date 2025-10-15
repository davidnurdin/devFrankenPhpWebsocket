# API Queue Counter - FrankenPHP WebSocket Extension

## Vue d'ensemble

Le système **Queue Counter** permet de tracer et stocker les messages envoyés à chaque client WebSocket. Il offre un compteur par client et une queue des derniers messages envoyés, avec gestion automatique de l'expiration.

## Fonctionnalités

- **Compteur par client** : Chaque client a son propre compteur qui s'incrémente à chaque message envoyé
- **Queue des derniers messages** : Stockage des N derniers messages ou messages des X dernières minutes
- **Activation optionnelle** : Par défaut, rien n'est stocké (performance optimale)
- **Nettoyage automatique** : Suppression automatique des messages expirés
- **Support de tous les types d'envoi** : direct, tag, tagExpression, sendAll

## Fonctions disponibles

### `frankenphp_ws_enableQueueCounter(string $connectionId, int $maxMessages = 100, int $maxTimeSeconds = 3600): bool`

Active le compteur et la queue des messages pour un client spécifique.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion client
- `$maxMessages` (int, optionnel) : Nombre maximum de messages à conserver (0 = pas de limite)
- `$maxTimeSeconds` (int, optionnel) : Durée maximale de conservation en secondes (0 = pas de limite)

**Retour :**
- `bool` : `true` si activé avec succès, `false` sinon

**Comportement :**
- Active le compteur et la queue pour le client spécifié
- Les messages suivants seront automatiquement tracés
- Nettoyage automatique basé sur les limites configurées

### `frankenphp_ws_disableQueueCounter(string $connectionId): bool`

Désactive le compteur et la queue des messages pour un client spécifique.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion client

**Retour :**
- `bool` : `true` si désactivé avec succès, `false` sinon

**Comportement :**
- Désactive le traçage des messages pour ce client
- Les messages existants restent en queue jusqu'à expiration
- Aucun nouveau message ne sera tracé

### `frankenphp_ws_getClientMessageCounter(string $connectionId): int`

Retourne le compteur de messages pour un client spécifique.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion client

**Retour :**
- `int` : Nombre total de messages envoyés à ce client (depuis l'activation)

**Comportement :**
- Retourne 0 si le client n'a pas de compteur activé
- Le compteur s'incrémente à chaque message envoyé (tous types confondus)

### `frankenphp_ws_getClientMessageQueue(string $connectionId): array`

Retourne la queue des messages pour un client spécifique.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion client

**Retour :**
- `array` : Tableau des messages en queue (format : "ID:X|Route:Y|Time:Z|SendType:A|SendTarget:B|Data:C")

**Comportement :**
- Retourne un tableau vide si aucune queue n'est activée
- Chaque élément contient les informations du message
- Les messages sont triés par ordre chronologique

### `frankenphp_ws_clearClientMessageQueue(string $connectionId): bool`

Vide la queue des messages pour un client spécifique.

**Paramètres :**
- `$connectionId` (string) : ID de la connexion client

**Retour :**
- `bool` : `true` si vidé avec succès, `false` sinon

**Comportement :**
- Supprime tous les messages en queue pour ce client
- Le compteur reste inchangé
- La queue reste activée (nouveaux messages seront tracés)

## Exemples d'utilisation

### Activation basique

```php
// Activer la queue avec les paramètres par défaut (100 messages max, 1 heure max)
$success = frankenphp_ws_enableQueueCounter("client_123");

if ($success) {
    echo "Queue counter activé pour client_123\n";
}
```

### Configuration personnalisée

```php
// Activer avec 50 messages maximum et 30 minutes de rétention
$success = frankenphp_ws_enableQueueCounter("client_123", 50, 1800);

if ($success) {
    echo "Queue counter activé avec limites personnalisées\n";
}
```

### Configuration sans limite de messages

```php
// Activer avec seulement une limite de temps (2 heures)
$success = frankenphp_ws_enableQueueCounter("client_123", 0, 7200);

if ($success) {
    echo "Queue counter activé avec limite de temps uniquement\n";
}
```

### Configuration sans limite de temps

```php
// Activer avec seulement une limite de messages (200 messages)
$success = frankenphp_ws_enableQueueCounter("client_123", 200, 0);

if ($success) {
    echo "Queue counter activé avec limite de messages uniquement\n";
}
```

### Consultation du compteur

```php
// Obtenir le compteur de messages
$counter = frankenphp_ws_getClientMessageCounter("client_123");
echo "Messages envoyés à client_123 : " . $counter . "\n";
```

### Consultation de la queue

```php
// Obtenir la queue des messages
$queue = frankenphp_ws_getClientMessageQueue("client_123");

foreach ($queue as $messageData) {
    // Format: "ID:123|Route:/chat|Time:1640995200|SendType:direct|SendTarget:client_123|Data:Hello"
    echo "Message en queue : " . $messageData . "\n";
}

echo "Nombre de messages en queue : " . count($queue) . "\n";
```

### Gestion de la queue

```php
// Vider la queue
$success = frankenphp_ws_clearClientMessageQueue("client_123");
if ($success) {
    echo "Queue vidée avec succès\n";
}

// Désactiver la queue
$success = frankenphp_ws_disableQueueCounter("client_123");
if ($success) {
    echo "Queue counter désactivé\n";
}
```

### Gestion de plusieurs connexions

```php
$connections = ["client_1", "client_2", "client_3"];

// Activer la queue pour tous les clients
foreach ($connections as $clientId) {
    frankenphp_ws_enableQueueCounter($clientId, 100, 3600);
}

// Envoyer des messages (ils seront automatiquement tracés)
foreach ($connections as $clientId) {
    frankenphp_ws_send($clientId, "Message pour " . $clientId, "/chat");
}

// Consulter les compteurs
foreach ($connections as $clientId) {
    $counter = frankenphp_ws_getClientMessageCounter($clientId);
    echo "Client $clientId : $counter messages\n";
}
```

## Types de messages tracés

Le système trace automatiquement tous les types de messages envoyés :

### Messages directs
```php
frankenphp_ws_send("client_123", "Hello", "/chat");
// → Traced avec SendType: "direct", SendTarget: "client_123"
```

### Messages par tag
```php
frankenphp_ws_sendToTag("premium", "Message premium", "/chat");
// → Traced pour tous les clients avec le tag "premium"
// → SendType: "tag", SendTarget: "premium"
```

### Messages par expression de tags
```php
frankenphp_ws_sendToTagExpression("premium AND active", "Message ciblé", "/chat");
// → Traced pour tous les clients correspondant à l'expression
// → SendType: "tagExpression", SendTarget: "premium AND active"
```

### Messages en masse
```php
frankenphp_ws_sendAll("Message à tous", "/chat");
// → Traced pour tous les clients connectés
// → SendType: "all", SendTarget: "all"
```

## Format des messages en queue

Chaque message en queue est retourné sous forme de chaîne avec le format suivant :

```
ID:123|Route:/chat|Time:1640995200|SendType:direct|SendTarget:client_123|Data:Hello World
```

**Composants :**
- `ID` : Identifiant unique du message (incrémental par client)
- `Route` : Route WebSocket du message
- `Time` : Timestamp Unix du message
- `SendType` : Type d'envoi ("direct", "tag", "tagExpression", "all")
- `SendTarget` : Cible de l'envoi (clientID, tag, expression, ou "all")
- `Data` : Contenu du message

## API REST Admin

Toutes les fonctions sont également disponibles via l'API REST Admin de Caddy :

### Activer la queue counter
```bash
POST /frankenphp_ws/enableQueueCounter/{clientID}?maxMessages=100&maxTime=3600
```

### Désactiver la queue counter
```bash
POST /frankenphp_ws/disableQueueCounter/{clientID}
```

### Obtenir le compteur
```bash
GET /frankenphp_ws/getClientMessageCounter/{clientID}
```

### Obtenir la queue
```bash
GET /frankenphp_ws/getClientMessageQueue/{clientID}
```

### Vider la queue
```bash
POST /frankenphp_ws/clearClientMessageQueue/{clientID}
```

## Cas d'usage

### 1. Débogage et monitoring
```php
// Activer le traçage pour un client problématique
frankenphp_ws_enableQueueCounter("problematic_client", 1000, 7200);

// Envoyer des messages de test
frankenphp_ws_send("problematic_client", "Test message 1", "/debug");
frankenphp_ws_send("problematic_client", "Test message 2", "/debug");

// Analyser les messages envoyés
$queue = frankenphp_ws_getClientMessageQueue("problematic_client");
foreach ($queue as $message) {
    echo "Message envoyé : " . $message . "\n";
}
```

### 2. Système de retry
```php
// Activer la queue pour un client critique
frankenphp_ws_enableQueueCounter("critical_client", 50, 1800);

// Envoyer des messages importants
frankenphp_ws_send("critical_client", "Message critique 1", "/important");
frankenphp_ws_send("critical_client", "Message critique 2", "/important");

// Vérifier si le client a reçu les messages
$counter = frankenphp_ws_getClientMessageCounter("critical_client");
if ($counter > 0) {
    // Implémenter une logique de retry basée sur la queue
    $queue = frankenphp_ws_getClientMessageQueue("critical_client");
    // ... logique de retry ...
}
```

### 3. Audit et logs
```php
// Activer le traçage pour tous les clients premium
$premiumClients = frankenphp_ws_getClientsByTag("premium");
foreach ($premiumClients as $clientId) {
    frankenphp_ws_enableQueueCounter($clientId, 200, 3600);
}

// Envoyer des messages premium
frankenphp_ws_sendToTag("premium", "Message premium", "/vip");

// Générer un rapport d'audit
foreach ($premiumClients as $clientId) {
    $counter = frankenphp_ws_getClientMessageCounter($clientId);
    $queue = frankenphp_ws_getClientMessageQueue($clientId);
    echo "Client $clientId : $counter messages, " . count($queue) . " en queue\n";
}
```

## Performance et limitations

### Performance
- **Par défaut** : Aucun impact sur les performances (rien n'est tracé)
- **Activé** : Impact minimal, traçage en mémoire uniquement
- **Nettoyage automatique** : Suppression automatique des messages expirés

### Limitations
- **Mémoire** : Les messages sont stockés en mémoire (pas de persistance)
- **Redémarrage** : Les données sont perdues lors du redémarrage du serveur
- **Concurrence** : Thread-safe, support des accès concurrents

### Recommandations
- **Activer uniquement si nécessaire** : Pour éviter l'impact sur les performances
- **Limiter la rétention** : Utiliser des limites raisonnables (maxMessages, maxTime)
- **Nettoyer régulièrement** : Vider les queues inutiles avec `clearClientMessageQueue`
- **Monitoring** : Surveiller l'utilisation mémoire si beaucoup de clients ont la queue activée

## Intégration avec d'autres APIs

### Avec le système de ping
```php
// Activer le ping et la queue pour un client
frankenphp_ws_enablePing("client_123", 5000); // Ping toutes les 5 secondes
frankenphp_ws_enableQueueCounter("client_123", 100, 3600); // Queue 100 messages, 1 heure

// Le ping sera aussi tracé dans la queue
$queue = frankenphp_ws_getClientMessageQueue("client_123");
// Contiendra les pings automatiques + les messages envoyés
```

### Avec le système de tags
```php
// Tagger un client et activer la queue
frankenphp_ws_tagClient("client_123", "premium");
frankenphp_ws_enableQueueCounter("client_123", 50, 1800);

// Envoyer des messages par tag (sera tracé pour ce client)
frankenphp_ws_sendToTag("premium", "Message premium", "/vip");
```

### Avec le système de stockage
```php
// Stocker des informations et activer la queue
frankenphp_ws_setStoredInformation("client_123", "queue_enabled", "true");
frankenphp_ws_enableQueueCounter("client_123", 100, 3600);

// Vérifier si la queue est activée
$queueEnabled = frankenphp_ws_getStoredInformation("client_123", "queue_enabled");
if ($queueEnabled === "true") {
    $counter = frankenphp_ws_getClientMessageCounter("client_123");
    echo "Queue activée, compteur : $counter\n";
}
```
