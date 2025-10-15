<?php

/** @generate-class-entries */

function frankenphp_ws_getClients(?string $route = null): array {}

function frankenphp_ws_getClientsCount(?string $route = null): int {}

function frankenphp_ws_send(string $connectionId, string $data, ?string $route = null): void {}

function frankenphp_ws_sendAll(string $data, ?string $route = null): int {}

function frankenphp_ws_killConnection(string $connectionId): bool {}

function frankenphp_ws_getClientPingTime(string $connectionId): int {}

function frankenphp_ws_enablePing(string $connectionId, int $intervalMs = 0): bool {}

function frankenphp_ws_disablePing(string $connectionId): bool {}

// ===== FONCTIONS POUR LA GESTION DE LA QUEUE COUNTER =====

function frankenphp_ws_enableQueueCounter(string $connectionId, int $maxMessages = 100, int $maxTimeSeconds = 3600): bool {}
function frankenphp_ws_disableQueueCounter(string $connectionId): bool {}
function frankenphp_ws_getClientMessageCounter(string $connectionId): int {}
function frankenphp_ws_getClientMessageQueue(string $connectionId): array {}
function frankenphp_ws_clearClientMessageQueue(string $connectionId): bool {}

function frankenphp_ws_tagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_untagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_clearTagClient(string $connectionId): void {}

function frankenphp_ws_getTags(): array {}

function frankenphp_ws_getClientsByTag(string $tag): array {}

function frankenphp_ws_getTagCount(string $tag): int {}

function frankenphp_ws_sendToTag(string $tag, string $data, ?string $route = null): void {}

function frankenphp_ws_setStoredInformation(string $connectionId, string $key, string $value): void {}

function frankenphp_ws_getStoredInformation(string $connectionId, string $key): string {}

function frankenphp_ws_deleteStoredInformation(string $connectionId, string $key): void {}

function frankenphp_ws_clearStoredInformation(string $connectionId): void {}

function frankenphp_ws_hasStoredInformation(string $connectionId, string $key): bool {}

function frankenphp_ws_listStoredInformationKeys(string $connectionId): array {}

function frankenphp_ws_sendToTagExpression(string $expression, string $data, ?string $route = null): void {}

function frankenphp_ws_getClientsByTagExpression(string $expression): array {}

function frankenphp_ws_listRoutes(): array {}

// ===== Ghost connection management =====
function frankenphp_ws_activateGhost(string $connectionId): bool {}
function frankenphp_ws_releaseGhost(string $connectionId): bool {}
function frankenphp_ws_isGhost(string $connectionId): bool {}

function frankenphp_ws_renameConnection(string $currentId, string $newId): bool {}

// ===== Global information (in-memory, thread-safe, with expiration) =====
// expireSeconds: 0 for infinite; >0 means N seconds
function frankenphp_ws_global_set(string $key, string $value, int $expireSeconds = 0): void {}
function frankenphp_ws_global_get(string $key): string {}
function frankenphp_ws_global_has(string $key): bool {}
function frankenphp_ws_global_delete(string $key): bool {}

// ===== Stored Information search =====
// Retourne la liste d'IDs correspondant à key/op/value, filtrable par route
/**
 * Constantes PHP pour les opérateurs de recherche (déclarées en C dans MINIT):
 * - FRANKENPHP_WS_OP_EQ         => 'eq'
 * - FRANKENPHP_WS_OP_NEQ        => 'neq'
 * - FRANKENPHP_WS_OP_PREFIX     => 'prefix'
 * - FRANKENPHP_WS_OP_SUFFIX     => 'suffix'
 * - FRANKENPHP_WS_OP_CONTAINS   => 'contains'
 * - FRANKENPHP_WS_OP_IEQ        => 'ieq'
 * - FRANKENPHP_WS_OP_IPREFIX    => 'iprefix'
 * - FRANKENPHP_WS_OP_ISUFFIX    => 'isuffix'
 * - FRANKENPHP_WS_OP_ICONTAINS  => 'icontains'
 * - FRANKENPHP_WS_OP_REGEX      => 'regex'
 * 
 * Ces constantes sont enregistrées via REGISTER_STRING_CONSTANT dans la fonction MINIT
 */
function frankenphp_ws_searchStoredInformation(string $key, string $op, string $value, ?string $route = null): array {}

// ===== Constantes de l'extension (déclarées en C) =====
// Note: Ces constantes sont définies dans le code C de l'extension et disponibles
// automatiquement en PHP une fois l'extension compilée et chargée.

/** @var string */
const FRANKENPHP_WS_OP_EQ = 'eq';

/** @var string */
const FRANKENPHP_WS_OP_NEQ = 'neq';

/** @var string */
const FRANKENPHP_WS_OP_PREFIX = 'prefix';

/** @var string */
const FRANKENPHP_WS_OP_SUFFIX = 'suffix';

/** @var string */
const FRANKENPHP_WS_OP_CONTAINS = 'contains';

/** @var string */
const FRANKENPHP_WS_OP_IEQ = 'ieq';

/** @var string */
const FRANKENPHP_WS_OP_IPREFIX = 'iprefix';

/** @var string */
const FRANKENPHP_WS_OP_ISUFFIX = 'isuffix';

/** @var string */
const FRANKENPHP_WS_OP_ICONTAINS = 'icontains';

/** @var string */
const FRANKENPHP_WS_OP_REGEX = 'regex';

