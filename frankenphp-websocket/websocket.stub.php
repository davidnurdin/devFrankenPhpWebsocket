<?php

/** @generate-class-entries */

function frankenphp_ws_getClients(?string $route = null): array {}

function frankenphp_ws_send(string $connectionId, string $data, ?string $route = null): void {}

function frankenphp_ws_tagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_untagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_clearTagClient(string $connectionId): void {}

function frankenphp_ws_getTags(): array {}

function frankenphp_ws_getClientsByTag(string $tag): array {}

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
 * Constantes PHP pour les opérateurs de recherche:
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
 */
function frankenphp_ws_searchStoredInformation(string $key, string $op, string $value, ?string $route = null): array {}

