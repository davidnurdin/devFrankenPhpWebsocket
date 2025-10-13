# FrankenPHP WebSocket Extension

This project demonstrates how to use the **Go extension** in **FrankenPHP** to implement a **native WebSocket server** directly integrated into FrankenPHP — without relying on an external proxy or process.

🎥 **Video demonstration:** [Watch on YouTube](https://www.youtube.com/watch?v=z9sluTIjgwQ)

---

## 🧩 Overview

This extension extends FrankenPHP with native WebSocket capabilities written in **Go**, seamlessly bridging PHP and Go through the FrankenPHP runtime.

It allows PHP developers to manage WebSocket connections **directly from PHP code**, while leveraging the high-performance concurrency model of Go.

---

## ⚙️ Features

- Native WebSocket server embedded within FrankenPHP  
- Fully event-driven and lightweight  
- Direct PHP functions to interact with connected clients  
- No external dependencies or additional WebSocket daemons required  

---

## 🧠 Available PHP Functions

The extension exposes multiple PHP functions for WebSocket management:

### Connection Management
- `frankenphp_ws_getClients()` - Returns the list of currently connected WebSocket clients
- `frankenphp_ws_send($clientId, $message)` - Sends a message to a specific connected client
- `frankenphp_ws_renameConnection($currentId, $newId)` - Renames a WebSocket connection ID while preserving all associated data

### Tag Management
- `frankenphp_ws_tagClient($connectionId, $tag)` - Tags a client with a specific tag
- `frankenphp_ws_untagClient($connectionId, $tag)` - Removes a tag from a client
- `frankenphp_ws_clearTagClient($connectionId)` - Removes all tags from a client
- `frankenphp_ws_getTags()` - Returns all available tags
- `frankenphp_ws_getClientsByTag($tag)` - Returns clients with a specific tag
- `frankenphp_ws_sendToTag($tag, $data)` - Sends a message to all clients with a tag

### Information Storage
- `frankenphp_ws_setStoredInformation($connectionId, $key, $value)` - Stores information for a connection
- `frankenphp_ws_getStoredInformation($connectionId, $key)` - Retrieves stored information
- `frankenphp_ws_deleteStoredInformation($connectionId, $key)` - Deletes specific information
- `frankenphp_ws_clearStoredInformation($connectionId)` - Clears all information for a connection
- `frankenphp_ws_hasStoredInformation($connectionId, $key)` - Checks if information exists
- `frankenphp_ws_listStoredInformationKeys($connectionId)` - Lists all stored information keys

### Tag Logic (NEW!)
- `frankenphp_ws_sendToTagExpression($expression, $data)` - Sends message with boolean tag logic
- `frankenphp_ws_getClientsByTagExpression($expression)` - Gets clients matching tag expression

**Tag Logic Operators:** `&` (AND), `|` (OR), `!` (NOT), `()` (parentheses), `*` (wildcard)

📖 **Detailed documentation:** 
- [STORED_INFORMATION_API.md](STORED_INFORMATION_API.md)
- [TAG_LOGIC_API.md](TAG_LOGIC_API.md)
- [CONNECTION_MANAGEMENT_API.md](CONNECTION_MANAGEMENT_API.md)

---

## 🚀 Example Usage (PHP)

### Basic WebSocket Communication
```php
<?php

$clients = frankenphp_ws_getClients();

foreach ($clients as $client) {
    frankenphp_ws_send($client, json_encode(['event' => 'ping', 'time' => time()]));
}
```

### Connection Renaming
```php
<?php

// Rename a connection while preserving all data
$success = frankenphp_ws_renameConnection('temp_connection_123', 'user_456');

if ($success) {
    // All tags, stored information, and routes are preserved
    frankenphp_ws_send('user_456', json_encode(['message' => 'Connection renamed successfully']));
}
```

### Information Storage Example
```php
<?php

// Store user information when they connect
$connectionId = 'client_12345';
frankenphp_ws_setStoredInformation($connectionId, 'user_id', '12345');
frankenphp_ws_setStoredInformation($connectionId, 'username', 'john_doe');
frankenphp_ws_setStoredInformation($connectionId, 'language', 'fr');

// Retrieve and use stored information
if (frankenphp_ws_hasStoredInformation($connectionId, 'user_id')) {
    $userId = frankenphp_ws_getStoredInformation($connectionId, 'user_id');
    $username = frankenphp_ws_getStoredInformation($connectionId, 'username');
    
    // Send personalized message
    frankenphp_ws_send($connectionId, json_encode([
        'message' => "Welcome back, $username!",
        'user_id' => $userId
    ]));
}

// List all stored information keys
$keys = frankenphp_ws_listStoredInformationKeys($connectionId);
foreach ($keys as $key) {
    $value = frankenphp_ws_getStoredInformation($connectionId, $key);
    echo "Stored: $key = $value\n";
}
```

### Tag Management Example
```php
<?php

// Tag clients by their role
frankenphp_ws_tagClient($connectionId, 'premium_user');
frankenphp_ws_tagClient($connectionId, 'french_speaker');

// Send message to all premium users
$premiumClients = frankenphp_ws_getClientsByTag('premium_user');
foreach ($premiumClients as $client) {
    frankenphp_ws_send($client, json_encode(['event' => 'premium_offer']));
}

// Or send directly to all clients with a tag
frankenphp_ws_sendToTag('french_speaker', json_encode(['message' => 'Bonjour!']));
```

### Tag Logic Example
```php
<?php

// Send to men from Grenoble (AND logic)
frankenphp_ws_sendToTagExpression('grenoble&homme', json_encode([
    'message' => 'Événement réservé aux hommes de Grenoble'
]));

// Send to premium or VIP users (OR logic)
frankenphp_ws_sendToTagExpression('premium|vip', json_encode([
    'message' => 'Contenu exclusif'
]));

// Send to authenticated users who are not banned (NOT logic)
frankenphp_ws_sendToTagExpression('authenticated&!banned', json_encode([
    'message' => 'Notification importante'
]));

// Complex expression with parentheses
frankenphp_ws_sendToTagExpression('(premium|vip)&(grenoble|lyon)&!banned', json_encode([
    'message' => 'Offre spéciale locale'
]));

// Get clients matching expression
$clients = frankenphp_ws_getClientsByTagExpression('admin&!test');
foreach ($clients as $clientId) {
    echo "Admin client: $clientId\n";
}

// Wildcard expressions (NEW!)
frankenphp_ws_sendToTagExpression('group_*', json_encode([
    'message' => 'Message to all groups'
]));

frankenphp_ws_sendToTagExpression('*admin', json_encode([
    'message' => 'Message to all admins'
]));

frankenphp_ws_sendToTagExpression('group_*&!banned', json_encode([
    'message' => 'Message to all groups not banned'
]));
```

