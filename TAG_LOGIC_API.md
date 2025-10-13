# FrankenPHP WebSocket - API de Logique de Tags

Cette documentation décrit les nouvelles fonctions de logique de tags pour les connexions WebSocket dans FrankenPHP, permettant d'envoyer des messages avec des expressions booléennes complexes.

## Vue d'ensemble

Le système de logique de tags permet d'utiliser des expressions booléennes pour cibler des clients WebSocket avec des critères complexes. Vous pouvez combiner plusieurs tags avec des opérateurs logiques pour créer des conditions précises.

## Opérateurs supportés

### Opérateurs de base
- **`&`** - ET (AND) : Les deux conditions doivent être vraies
- **`|`** - OU (OR) : Au moins une condition doit être vraie
- **`!`** - NON (NOT) : Inverse la condition
- **`()`** - Parenthèses : Contrôlent la priorité des opérations
- **`*`** - Wildcard : Matche n'importe quelle séquence de caractères

### Priorité des opérateurs
1. Parenthèses `()`
2. Négation `!`
3. ET `&`
4. OU `|`
5. Wildcard `*` (évalué en premier pour les patterns)

## Fonctions disponibles

### `frankenphp_ws_sendToTagExpression(string $expression, string $data): void`

Envoie un message à tous les clients correspondant à une expression de tags.

**Paramètres :**
- `$expression` (string) : Expression booléenne des tags
- `$data` (string) : Message à envoyer

**Exemples d'expressions :**
```php
// Clients ayant le tag "grenoble" ET le tag "homme"
frankenphp_ws_sendToTagExpression('grenoble&homme', $message);

// Clients ayant le tag "grenoble" OU le tag "lyon"
frankenphp_ws_sendToTagExpression('grenoble|lyon', $message);

// Clients ayant le tag "admin" mais PAS le tag "test"
frankenphp_ws_sendToTagExpression('admin&!test', $message);

// Clients ayant (grenoble OU lyon) ET homme
frankenphp_ws_sendToTagExpression('(grenoble|lyon)&homme', $message);

// Clients ayant grenoble OU (lyon ET femme)
frankenphp_ws_sendToTagExpression('grenoble|(lyon&femme)', $message);

// Wildcard expressions (NEW!)
frankenphp_ws_sendToTagExpression('group_*', $message);              // Tous les groupes
frankenphp_ws_sendToTagExpression('*admin', $message);               // Tous les admins
frankenphp_ws_sendToTagExpression('region_*', $message);             // Toutes les régions

// Wildcards with boolean logic
frankenphp_ws_sendToTagExpression('group_*&!banned', $message);      // Groupes non bannis
frankenphp_ws_sendToTagExpression('*admin|*moderator', $message);    // Admins ou modérateurs
frankenphp_ws_sendToTagExpression('(group_*|team_*)&active', $message); // Groupes/équipes actifs
```

### `frankenphp_ws_getClientsByTagExpression(string $expression): array`

Retourne la liste des clients correspondant à une expression de tags.

**Paramètres :**
- `$expression` (string) : Expression booléenne des tags

**Retour :**
- (array) : Liste des IDs de connexion correspondants

**Exemple :**
```php
// Obtenir tous les clients hommes de Grenoble
$clients = frankenphp_ws_getClientsByTagExpression('grenoble&homme');

foreach ($clients as $clientId) {
    echo "Client homme de Grenoble: $clientId\n";
}
```

## Wildcards

### Support des wildcards
Le système supporte l'opérateur wildcard `*` qui peut matcher n'importe quelle séquence de caractères dans un nom de tag.

**Patterns supportés :**
- **`prefix_*`** : Matche tous les tags commençant par "prefix_"
- **`*suffix`** : Matche tous les tags finissant par "suffix"
- **`prefix_*_suffix`** : Matche les tags avec un pattern au milieu
- **`*`** : Matche tous les tags (attention à la performance)

### Exemples de wildcards
```php
// Tous les groupes (group_dev, group_marketing, group_sales, etc.)
frankenphp_ws_sendToTagExpression('group_*', $message);

// Tous les rôles admin (super_admin, team_admin, site_admin, etc.)
frankenphp_ws_sendToTagExpression('*admin', $message);

// Toutes les régions (region_france, region_spain, region_italy, etc.)
frankenphp_ws_sendToTagExpression('region_*', $message);

// Tous les niveaux d'accès (level_1, level_2, level_premium, etc.)
frankenphp_ws_sendToTagExpression('level_*', $message);
```

### Wildcards avec logique booléenne
```php
// Tous les groupes non bannis
frankenphp_ws_sendToTagExpression('group_*&!banned', $message);

// Tous les admins ou modérateurs
frankenphp_ws_sendToTagExpression('*admin|*moderator', $message);

// Groupes ou équipes actifs
frankenphp_ws_sendToTagExpression('(group_*|team_*)&active', $message);

// Tous les utilisateurs de niveau premium ou VIP, non bannis
frankenphp_ws_sendToTagExpression('(*premium|*vip)&!banned', $message);
```

## Exemples d'utilisation

### 1. Ciblage géographique et démographique

```php
// Envoyer aux hommes de Grenoble
frankenphp_ws_sendToTagExpression('grenoble&homme', json_encode([
    'message' => 'Événement réservé aux hommes de Grenoble',
    'event' => 'local_meetup'
]));

// Envoyer aux femmes de Lyon
frankenphp_ws_sendToTagExpression('lyon&femme', json_encode([
    'message' => 'Événement réservé aux femmes de Lyon',
    'event' => 'local_meetup'
]));

// Envoyer à tous les habitants de Grenoble ou Lyon
frankenphp_ws_sendToTagExpression('grenoble|lyon', json_encode([
    'message' => 'Actualités locales',
    'event' => 'local_news'
]));
```

### 2. Gestion des rôles et permissions

```php
// Envoyer aux administrateurs authentifiés (pas en test)
frankenphp_ws_sendToTagExpression('admin&authenticated&!test', json_encode([
    'message' => 'Mise à jour système',
    'event' => 'system_update'
]));

// Envoyer aux utilisateurs premium ou VIP
frankenphp_ws_sendToTagExpression('premium|vip', json_encode([
    'message' => 'Contenu exclusif',
    'event' => 'premium_content'
]));

// Envoyer aux utilisateurs non-bannis
frankenphp_ws_sendToTagExpression('!banned', json_encode([
    'message' => 'Nouvelle fonctionnalité',
    'event' => 'feature_announcement'
]));
```

### 3. Ciblage par intérêts et préférences

```php
// Envoyer aux utilisateurs intéressés par la technologie ET le sport
frankenphp_ws_sendToTagExpression('tech&sport', json_encode([
    'message' => 'Nouvelle app de sport connecté',
    'event' => 'product_launch'
]));

// Envoyer aux utilisateurs francophones OU anglophones
frankenphp_ws_sendToTagExpression('french|english', json_encode([
    'message' => 'Contenu multilingue',
    'event' => 'multilingual_content'
]));

// Envoyer aux utilisateurs avec notifications activées mais pas en mode silencieux
frankenphp_ws_sendToTagExpression('notifications&!silent', json_encode([
    'message' => 'Notification importante',
    'event' => 'important_notification'
]));
```

### 4. Expressions complexes avec parenthèses

```php
// Envoyer aux utilisateurs (premium OU vip) ET (français OU anglais)
frankenphp_ws_sendToTagExpression('(premium|vip)&(french|english)', json_encode([
    'message' => 'Offre spéciale multilingue',
    'event' => 'special_offer'
]));

// Envoyer aux utilisateurs de (Grenoble OU Lyon) ET (homme OU femme) mais pas bannis
frankenphp_ws_sendToTagExpression('(grenoble|lyon)&(homme|femme)&!banned', json_encode([
    'message' => 'Événement local ouvert à tous',
    'event' => 'local_event'
]));

// Envoyer aux administrateurs OU (utilisateurs premium ET authentifiés)
frankenphp_ws_sendToTagExpression('admin|(premium&authenticated)', json_encode([
    'message' => 'Accès prioritaire',
    'event' => 'priority_access'
]));
```

### 5. Utilisation avec les informations stockées

```php
// Combiner tags et informations stockées
$clients = frankenphp_ws_getClientsByTagExpression('premium&!banned');

foreach ($clients as $clientId) {
    // Vérifier les informations stockées
    $age = frankenphp_ws_getStoredInformation($clientId, 'age');
    $interests = frankenphp_ws_getStoredInformation($clientId, 'interests');
    
    if ($age && intval($age) >= 18) {
        $interestsArray = json_decode($interests, true);
        if (in_array('technology', $interestsArray)) {
            frankenphp_ws_send($clientId, json_encode([
                'message' => 'Contenu tech pour adultes',
                'event' => 'adult_tech_content'
            ]));
        }
    }
}
```

## Cas d'usage avancés

### 1. Système de notifications intelligentes

```php
function sendSmartNotification($message, $priority = 'normal') {
    switch ($priority) {
        case 'urgent':
            // Tous les utilisateurs en ligne et non bannis
            frankenphp_ws_sendToTagExpression('online&!banned', $message);
            break;
            
        case 'premium':
            // Utilisateurs premium authentifiés
            frankenphp_ws_sendToTagExpression('premium&authenticated&!silent', $message);
            break;
            
        case 'local':
            // Utilisateurs d'une région spécifique
            frankenphp_ws_sendToTagExpression('(grenoble|lyon|paris)&!banned', $message);
            break;
            
        default:
            // Utilisateurs authentifiés avec notifications activées
            frankenphp_ws_sendToTagExpression('authenticated&notifications&!silent', $message);
    }
}
```

### 2. Système de chat par groupes

```php
function sendToGroup($groupId, $message) {
    // Envoyer aux membres du groupe qui sont en ligne et non bannis
    frankenphp_ws_sendToTagExpression("group_{$groupId}&online&!banned", $message);
}

function sendToMultipleGroups($groupIds, $message) {
    // Construire l'expression pour plusieurs groupes
    $expression = '';
    foreach ($groupIds as $index => $groupId) {
        if ($index > 0) {
            $expression .= '|';
        }
        $expression .= "group_{$groupId}";
    }
    $expression .= '&online&!banned';
    
    frankenphp_ws_sendToTagExpression($expression, $message);
}
```

### 3. Système de recommandations

```php
function sendRecommendations($clientId) {
    // Obtenir les tags du client
    $tags = frankenphp_ws_getClientTags($clientId);
    
    if (in_array('music', $tags)) {
        // Envoyer aux autres amateurs de musique
        frankenphp_ws_sendToTagExpression('music&!banned', json_encode([
            'type' => 'recommendation',
            'message' => 'Nouvelle recommandation musicale',
            'from' => $clientId
        ]));
    }
    
    if (in_array('sport', $tags)) {
        // Envoyer aux sportifs de la même région
        $region = frankenphp_ws_getStoredInformation($clientId, 'region');
        if ($region) {
            frankenphp_ws_sendToTagExpression("sport&region_{$region}&!banned", json_encode([
                'type' => 'recommendation',
                'message' => 'Événement sportif local',
                'from' => $clientId
            ]));
        }
    }
}
```

## Compatibilité CLI et Web

Comme toutes les autres fonctions WebSocket, la logique de tags fonctionne à la fois :
- **En mode CLI** : Les appels sont redirigés vers l'API admin de Caddy via HTTP
- **En mode Web** : Les appels sont traités directement par le serveur WebSocket

## Endpoints API Admin

Les fonctions utilisent également des endpoints API admin pour le mode CLI :

- `POST /frankenphp_ws/sendToTagExpression/{expression}` : Envoyer un message avec expression
- `GET /frankenphp_ws/getClientsByTagExpression/{expression}` : Récupérer les clients avec expression

## Limitations et bonnes pratiques

### Limitations
- Les expressions sont limitées à 1000 caractères
- Les noms de tags ne peuvent contenir que des lettres, chiffres et underscores
- Les parenthèses doivent être correctement équilibrées
- Pas de support pour les expressions imbriquées trop complexes

### Bonnes pratiques
1. **Utilisez des noms de tags courts et clairs** : `admin`, `premium`, `grenoble`
2. **Évitez les expressions trop complexes** : Préférez plusieurs appels simples
3. **Testez vos expressions** : Utilisez `getClientsByTagExpression()` pour vérifier
4. **Cachez les résultats** : Pour les expressions complexes fréquemment utilisées
5. **Documentez vos expressions** : Commentez les expressions complexes

### Exemples d'expressions valides
```php
// Simple
'admin'
'premium'
'grenoble'

// Avec opérateurs
'admin&premium'
'grenoble|lyon'
'!banned'

// Avec parenthèses
'(admin|premium)&!banned'
'(grenoble|lyon)&(homme|femme)'
'admin|(premium&authenticated)'
```

### Exemples d'expressions invalides
```php
// Parenthèses non équilibrées
'admin&(premium'        // Manque )
'(admin&premium'        // Manque )

// Caractères non autorisés
'admin-user'            // Tiret non autorisé
'admin@premium'         // @ non autorisé
'admin space'           // Espace non autorisé

// Expression trop complexe
'admin&premium&vip&authenticated&!banned&!silent&online&active&verified'
```

## Performance

- Les expressions sont évaluées en temps réel pour chaque client
- Utilisez des expressions simples pour de meilleures performances
- Pour des ciblages complexes fréquents, considérez l'utilisation de tags composites
- Les parenthèses peuvent améliorer la performance en évitant des évaluations inutiles

Cette API de logique de tags vous permet de créer des systèmes de ciblage très sophistiqués pour vos applications WebSocket !
