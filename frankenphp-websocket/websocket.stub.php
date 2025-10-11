<?php

/** @generate-class-entries */

function frankenphp_ws_getClients(): array {}

function frankenphp_ws_send(string $connectionId, string $data): void {}

function frankenphp_ws_tagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_untagClient(string $connectionId, string $tag): void {}

function frankenphp_ws_clearTagClient(string $connectionId): void {}

function frankenphp_ws_getTags(): array {}

function frankenphp_ws_getClientsByTag(string $tag): array {}

function frankenphp_ws_sendToTag(string $tag, string $data): void {}