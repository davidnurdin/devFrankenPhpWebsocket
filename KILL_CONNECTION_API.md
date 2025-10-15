# API de Fermeture de Connexion WebSocket

Cette API permet de fermer immédiatement une connexion WebSocket spécifique en utilisant son ID de connexion.

## Fonction PHP

### `frankenphp_ws_killConnection(string $connectionId): bool`

Ferme immédiatement une connexion WebSocket spécifique.

**Paramètres :**
- `$connectionId` (string) : L'ID de la connexion WebSocket à fermer

**Retour :**
- `bool` : `true` si la connexion a été fermée avec succès, `false` si la connexion n'a pas été trouvée ou était déjà fermée

**Exemples d'utilisation :**

```php
// Fermer une connexion spécifique
$success = frankenphp_ws_killConnection("client_123");
if ($success) {
    echo "Connexion fermée avec succès\n";
} else {
    echo "Connexion non trouvée ou déjà fermée\n";
}

// Fermer une connexion après vérification
$connectionId = "user_456";
if (frankenphp_ws_hasStoredInformation($connectionId, "username")) {
    $username = frankenphp_ws_getStoredInformation($connectionId, "username");
    echo "Fermeture de la connexion pour l'utilisateur: $username\n";
    
    $success = frankenphp_ws_killConnection($connectionId);
    if ($success) {
        echo "Utilisateur déconnecté avec succès\n";
    }
}

// Fermer toutes les connexions d'un utilisateur spécifique
function disconnectUser($username) {
    $clients = frankenphp_ws_searchStoredInformation("username", "eq", $username);
    $disconnectedCount = 0;
    
    foreach ($clients as $connectionId) {
        if (frankenphp_ws_killConnection($connectionId)) {
            $disconnectedCount++;
        }
    }
    
    return $disconnectedCount;
}

$disconnectedCount = disconnectUser("john_doe");
echo "Déconnecté $disconnectedCount connexions pour john_doe\n";
```

## API Admin (CLI)

### Endpoint

**POST** `/frankenphp_ws/killConnection/{clientID}`

### Paramètres

- `clientID` (URL) : L'ID de la connexion WebSocket à fermer

### Réponses

**Succès :**
```json
{
  "clientID": "client_123",
  "success": true
}
```

**Échec (connexion non trouvée) :**
```json
{
  "clientID": "client_123",
  "success": false,
  "error": "Connection not found or already closed"
}
```

### Exemples d'utilisation

```bash
# Fermer une connexion spécifique
curl -X POST http://localhost:2019/frankenphp_ws/killConnection/client_123

# Fermer une connexion avec un ID complexe (nécessite l'encodage URL)
curl -X POST "http://localhost:2019/frankenphp_ws/killConnection/user%3A456%3Asession%3A789"
```

## Cas d'usage

### 1. Déconnexion forcée d'utilisateurs
```php
// Fonction pour déconnecter un utilisateur par son nom d'utilisateur
function forceDisconnectUser($username) {
    $clients = frankenphp_ws_searchStoredInformation("username", "eq", $username);
    $results = [];
    
    foreach ($clients as $connectionId) {
        $success = frankenphp_ws_killConnection($connectionId);
        $results[] = [
            'connectionId' => $connectionId,
            'success' => $success
        ];
    }
    
    return $results;
}

// Utilisation
$results = forceDisconnectUser("john_doe");
foreach ($results as $result) {
    if ($result['success']) {
        echo "Connexion {$result['connectionId']} fermée\n";
    } else {
        echo "Échec de fermeture pour {$result['connectionId']}\n";
    }
}
```

### 2. Nettoyage des connexions inactives
```php
// Fonction pour fermer les connexions inactives
function cleanupInactiveConnections($maxInactiveMinutes = 30) {
    $allClients = frankenphp_ws_getClients();
    $cutoffTime = time() - ($maxInactiveMinutes * 60);
    $cleanedCount = 0;
    
    foreach ($allClients as $connectionId) {
        // Vérifier la dernière activité (si stockée)
        if (frankenphp_ws_hasStoredInformation($connectionId, "lastActivity")) {
            $lastActivity = (int)frankenphp_ws_getStoredInformation($connectionId, "lastActivity");
            
            if ($lastActivity < $cutoffTime) {
                if (frankenphp_ws_killConnection($connectionId)) {
                    $cleanedCount++;
                    echo "Connexion inactive fermée: $connectionId\n";
                }
            }
        }
    }
    
    return $cleanedCount;
}

// Nettoyer les connexions inactives depuis plus de 30 minutes
$cleanedCount = cleanupInactiveConnections(30);
echo "Nettoyage terminé: $cleanedCount connexions fermées\n";
```

### 3. Gestion des violations de sécurité
```php
// Fonction pour déconnecter les utilisateurs suspects
function disconnectSuspiciousUsers() {
    $suspiciousClients = frankenphp_ws_searchStoredInformation("violationCount", "gte", "3");
    $disconnectedCount = 0;
    
    foreach ($suspiciousClients as $connectionId) {
        $username = frankenphp_ws_getStoredInformation($connectionId, "username");
        echo "Déconnexion de l'utilisateur suspect: $username\n";
        
        if (frankenphp_ws_killConnection($connectionId)) {
            $disconnectedCount++;
            
            // Envoyer une notification aux administrateurs
            $adminClients = frankenphp_ws_getClientsByTag("admin");
            foreach ($adminClients as $adminId) {
                frankenphp_ws_send($adminId, json_encode([
                    'type' => 'security_alert',
                    'message' => "Utilisateur $username déconnecté pour violations",
                    'connectionId' => $connectionId
                ]));
            }
        }
    }
    
    return $disconnectedCount;
}

// Déconnecter les utilisateurs suspects
$disconnectedCount = disconnectSuspiciousUsers();
echo "Sécurité: $disconnectedCount utilisateurs suspects déconnectés\n";
```

### 4. Maintenance du serveur
```php
// Fonction pour déconnecter tous les clients avant maintenance
function disconnectAllClientsForMaintenance($maintenanceMessage = "Maintenance programmée") {
    $allClients = frankenphp_ws_getClients();
    $disconnectedCount = 0;
    
    // Envoyer un message d'avertissement d'abord
    frankenphp_ws_sendAll(json_encode([
        'type' => 'maintenance_warning',
        'message' => $maintenanceMessage,
        'timestamp' => time()
    ]));
    
    // Attendre 5 secondes pour que les clients reçoivent le message
    sleep(5);
    
    // Fermer toutes les connexions
    foreach ($allClients as $connectionId) {
        if (frankenphp_ws_killConnection($connectionId)) {
            $disconnectedCount++;
        }
    }
    
    return $disconnectedCount;
}

// Déconnecter tous les clients pour maintenance
$disconnectedCount = disconnectAllClientsForMaintenance("Maintenance programmée dans 5 secondes");
echo "Maintenance: $disconnectedCount clients déconnectés\n";
```

### 5. Gestion des limites de connexions
```php
// Fonction pour fermer les connexions excédentaires
function enforceConnectionLimits($maxConnectionsPerUser = 3) {
    $allClients = frankenphp_ws_getClients();
    $userConnections = [];
    $disconnectedCount = 0;
    
    // Compter les connexions par utilisateur
    foreach ($allClients as $connectionId) {
        if (frankenphp_ws_hasStoredInformation($connectionId, "username")) {
            $username = frankenphp_ws_getStoredInformation($connectionId, "username");
            $userConnections[$username][] = $connectionId;
        }
    }
    
    // Fermer les connexions excédentaires
    foreach ($userConnections as $username => $connections) {
        if (count($connections) > $maxConnectionsPerUser) {
            $excessConnections = array_slice($connections, $maxConnectionsPerUser);
            
            foreach ($excessConnections as $connectionId) {
                if (frankenphp_ws_killConnection($connectionId)) {
                    $disconnectedCount++;
                    echo "Connexion excédentaire fermée pour $username: $connectionId\n";
                }
            }
        }
    }
    
    return $disconnectedCount;
}

// Appliquer les limites de connexions
$disconnectedCount = enforceConnectionLimits(3);
echo "Limites appliquées: $disconnectedCount connexions excédentaires fermées\n";
```

### 6. Déconnexion par route
```php
// Fonction pour fermer toutes les connexions d'une route spécifique
function disconnectAllClientsFromRoute($route) {
    $clients = frankenphp_ws_getClients($route);
    $disconnectedCount = 0;
    
    foreach ($clients as $connectionId) {
        if (frankenphp_ws_killConnection($connectionId)) {
            $disconnectedCount++;
        }
    }
    
    return $disconnectedCount;
}

// Fermer toutes les connexions du chat
$disconnectedCount = disconnectAllClientsFromRoute("/chat");
echo "Route /chat fermée: $disconnectedCount clients déconnectés\n";
```

## Compatibilité

- **Mode CLI** : Utilise l'API admin de Caddy via HTTP POST
- **Mode Worker** : Appel direct aux fonctions Go
- **Thread-safe** : Utilise des mutex pour la sécurité concurrente

## Notes importantes

- La fermeture est **immédiate** et **forcée**
- La connexion sera fermée même si le client est en train d'envoyer des données
- Les événements `beforeClose` et `close` seront déclenchés normalement
- Le nettoyage automatique des données (tags, informations stockées) se fera via les événements de fermeture
- La fonction retourne `false` si la connexion n'existe pas ou est déjà fermée

## Limitations

- Ne peut pas fermer des connexions par critères complexes (utilisez d'abord `searchStoredInformation` pour trouver les IDs)
- La fermeture est définitive - le client devra se reconnecter
- Les données en cours de transmission peuvent être perdues
- Ne fonctionne que sur les connexions WebSocket actives

## Sécurité

⚠️ **Attention** : Cette fonction doit être utilisée avec précaution car elle force la fermeture des connexions. Assurez-vous de :

- Vérifier les permissions avant de fermer des connexions
- Notifier les utilisateurs avant la fermeture quand c'est possible
- Logger les actions de fermeture pour audit
- Utiliser des délais appropriés pour les notifications de maintenance

## Différences avec les autres fonctions

| Fonction | Action | Cible |
|----------|--------|-------|
| `frankenphp_ws_killConnection` | Ferme immédiatement | Connexion spécifique |
| `frankenphp_ws_send` | Envoie un message | Connexion spécifique |
| `frankenphp_ws_sendAll` | Envoie un message | Toutes les connexions |
| `frankenphp_ws_sendToTag` | Envoie un message | Connexions avec un tag |
