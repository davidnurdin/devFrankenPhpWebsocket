# API des Connexions Fantômes WebSocket

## Vue d'ensemble

Le système de connexions fantômes permet de gérer des connexions WebSocket de manière spéciale. Quand une connexion est marquée comme "fantôme", les déconnexions normales sont ignorées et aucun événement de fermeture n'est déclenché. Cela permet de maintenir l'état de la connexion côté serveur même si le client se déconnecte temporairement.

## Fonctionnalités

### 1. Activation du mode fantôme
```php
bool frankenphp_ws_activateGhost(string $connectionId)
```

Marque une connexion comme fantôme. À partir de ce moment :
- Les déconnexions sont ignorées
- Aucun événement `beforeClose` ou `close` n'est déclenché
- La connexion reste "vivante" côté serveur
- Les données associées (tags, informations stockées, etc.) sont préservées

**Retour :** `true` si l'activation a réussi, `false` sinon.

### 2. Libération d'une connexion fantôme
```php
bool frankenphp_ws_releaseGhost(string $connectionId)
```

Libère une connexion fantôme et déclenche les événements de fermeture dans l'ordre suivant :
1. `ghostConnectionClose` - Nouvel événement spécifique aux connexions fantômes
2. `beforeClose` - Événement de pré-fermeture standard
3. `close` - Événement de fermeture final

**Retour :** `true` si la libération a réussi, `false` sinon.

### 3. Vérification du statut fantôme
```php
bool frankenphp_ws_isGhost(string $connectionId)
```

Vérifie si une connexion est actuellement en mode fantôme.

**Retour :** `true` si la connexion est fantôme, `false` sinon.

## Nouvel événement : ghostConnectionClose

Un nouvel événement a été ajouté pour gérer spécifiquement la libération des connexions fantômes :

```php
// Dans votre worker PHP
switch ($event['type']) {
    case 'ghostConnectionClose':
        // Gestion spécifique de la libération d'une connexion fantôme
        $connectionId = $event['connection'];
        $payload = $event['payload']; // "ghost_released"
        
        // Actions à effectuer avant la fermeture normale
        // Par exemple : sauvegarder l'état, notifier d'autres services
        break;
}
```

## Cas d'usage typiques

### 1. Gestion de reconnexion
```php
// Client se déconnecte temporairement
frankenphp_ws_activateGhost($connectionId);

// Client se reconnecte avec le même ID
// L'ancienne connexion fantôme est libérée
frankenphp_ws_releaseGhost($oldConnectionId);
```

### 2. Maintenance programmée
```php
// Avant maintenance
$clients = frankenphp_ws_getClients();
foreach ($clients as $clientId) {
    frankenphp_ws_activateGhost($clientId);
}

// Pendant la maintenance, les déconnexions sont ignorées

// Après maintenance
foreach ($clients as $clientId) {
    if (frankenphp_ws_isGhost($clientId)) {
        frankenphp_ws_releaseGhost($clientId);
    }
}
```

### 3. Opérations critiques
```php
// Pendant une opération critique
frankenphp_ws_activateGhost($connectionId);

// Effectuer l'opération...
// Les déconnexions accidentelles sont ignorées

// Une fois l'opération terminée
frankenphp_ws_releaseGhost($connectionId);
```

## Flux d'événements

### Connexion normale
```
open → message → beforeClose → close
```

### Connexion fantôme (déconnexion ignorée)
```
open → message → activateGhost() → (disconnect ignoré)
```

### Libération de connexion fantôme
```
ghostConnectionClose → beforeClose → close
```

## Endpoints Admin API

Les connexions fantômes sont également accessibles via l'API admin :

### Activer une connexion fantôme
```bash
POST /frankenphp_ws/activateGhost/{clientID}
```

### Libérer une connexion fantôme
```bash
POST /frankenphp_ws/releaseGhost/{clientID}
```

### Vérifier le statut fantôme
```bash
GET /frankenphp_ws/isGhost/{clientID}
```

## Exemple complet

```php
<?php

// Worker WebSocket avec gestion des connexions fantômes
while (true) {
    $event = frankenphp_ws_get_event();
    
    if (!$event) {
        usleep(10000);
        continue;
    }
    
    switch ($event['type']) {
        case 'open':
            echo "Nouvelle connexion: " . $event['connection'] . "\n";
            break;
            
        case 'message':
            $data = json_decode($event['payload'], true);
            
            if ($data['type'] === 'activate_ghost') {
                $success = frankenphp_ws_activateGhost($event['connection']);
                echo "Mode fantôme activé: " . ($success ? "OUI" : "NON") . "\n";
            }
            break;
            
        case 'ghostConnectionClose':
            echo "Connexion fantôme libérée: " . $event['connection'] . "\n";
            // Sauvegarder l'état, notifier d'autres services, etc.
            break;
            
        case 'beforeClose':
            echo "Préparation fermeture: " . $event['connection'] . "\n";
            // Nettoyage des ressources
            break;
            
        case 'close':
            echo "Connexion fermée: " . $event['connection'] . "\n";
            // Actions finales
            break;
    }
}
```

## Notes importantes

1. **Thread-safety** : Toutes les fonctions sont thread-safe et peuvent être appelées depuis plusieurs goroutines simultanément.

2. **Persistance** : Les connexions fantômes sont stockées en mémoire et ne survivent pas au redémarrage du serveur.

3. **Performance** : L'activation/désactivation du mode fantôme est une opération très rapide (O(1)).

4. **Compatibilité** : Le système est rétrocompatible et n'affecte pas les connexions normales.

5. **Nettoyage automatique** : Les connexions fantômes sont automatiquement nettoyées lors de la libération.

## Dépannage

### Problème : La connexion fantôme n'est pas activée
- Vérifiez que l'ID de connexion existe
- Vérifiez les logs pour les erreurs

### Problème : Les événements ne sont pas déclenchés lors de la libération
- Vérifiez que la connexion était bien en mode fantôme
- Vérifiez que le worker PHP gère l'événement `ghostConnectionClose`

### Problème : La connexion n'est pas libérée
- Vérifiez que `frankenphp_ws_releaseGhost()` retourne `true`
- Vérifiez les logs pour les erreurs de libération
