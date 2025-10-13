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