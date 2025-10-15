## Global Information API

API de stockage global clé/valeur en mémoire, thread-safe, avec expiration optionnelle. Fonctionne en mode serveur (Caddy/FrankenPHP) et en mode CLI.

### PHP API

- `frankenphp_ws_global_set(string $key, string $value, int $expireSeconds = 0): void`
  - Enregistre la valeur pour `$key`. `expireSeconds = 0` signifie sans expiration (infini). `> 0` signifie expiration dans N secondes.

- `frankenphp_ws_global_get(string $key): string`
  - Retourne la valeur associée à `$key`, ou une chaîne vide si absente/expirée.

- `frankenphp_ws_global_has(string $key): bool`
  - Retourne `true` si la clé existe et n'est pas expirée, `false` sinon.

- `frankenphp_ws_global_delete(string $key): bool`
  - Supprime la clé. Retourne `true` si la clé existait, `false` sinon.

Notes:
- Les valeurs sont des chaînes UTF-8. Pour stocker des objets/arrays: `json_encode`/`json_decode` côté PHP.
- L'expiration est vérifiée à la lecture/has; lorsqu'une clé a expiré, elle est supprimée à la volée.

### Exemple PHP

```php
frankenphp_ws_global_set('maintenance', 'on', 60); // expire dans 60s

if (frankenphp_ws_global_has('maintenance')) {
    $mode = frankenphp_ws_global_get('maintenance');
    // ...
}

$deleted = frankenphp_ws_global_delete('maintenance');
```

### Comportement CLI vs Serveur

- En mode serveur (FrankenPHP/Caddy): les appels accèdent directement à la mémoire partagée du processus et sont protégés par `RWMutex`.
- En mode CLI: les appels passent par l'API Admin HTTP locale de Caddy pour atteindre le même état en mémoire du serveur.

### Endpoints Admin HTTP

Ces routes sont exposées sur l'API Admin (par défaut `http://localhost:2019`). Elles sont principalement destinées au mode CLI et au diagnostic.

- `POST /frankenphp_ws/global/set/{key}?exp=N`  (body: valeur, `Content-Type: text/plain`)
  - Définit la clé. `exp=N` (secondes) optionnel, `0` ou absent = infini.

- `GET /frankenphp_ws/global/get/{key}`
  - 200 + body=valeur si présente; 404 si absente/expirée.

- `GET /frankenphp_ws/global/has/{key}`
  - 200 si présente et non expirée; 404 sinon.

- `DELETE /frankenphp_ws/global/delete/{key}`
  - 200 si supprimée; 404 si absente.

### Concurrence

- Accès protégés par `globalInfoMutex (sync.RWMutex)`.
- Lectures concurrentes autorisées; écritures/suppressions sérialisées.


