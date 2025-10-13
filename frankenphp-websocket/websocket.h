#ifndef _WEBSOCKET_H
#define _WEBSOCKET_H

#include <php.h>

extern zend_module_entry ext_module_entry;

// Append a client ID into the specified PHP array
void frankenphp_ws_addClient(zval* array, const char* id);

// Get clients and populate the array (new signature)
void frankenphp_ws_getClients(void* array, char* route);

// Get all active routes and populate the array
void frankenphp_ws_listRoutes(void* array);

// Send data to a specific client by connection ID (matches cgo export signature)
void frankenphp_ws_send(char* connectionId, char* data, int data_len, char* route);

// Tag management functions
void frankenphp_ws_tagClient(char* connectionId, char* tag);
void frankenphp_ws_untagClient(char* connectionId, char* tag);
void frankenphp_ws_clearTagClient(char* connectionId);

// Tag query and broadcast functions
void frankenphp_ws_getTags(void* array);
void frankenphp_ws_getClientsByTag(void* array, char* tag);
void frankenphp_ws_sendToTag(char* tag, char* data, int data_len, char* route);

// Stored information management functions
void frankenphp_ws_setStoredInformation(char* connectionId, char* key, char* value);
char* frankenphp_ws_getStoredInformation(char* connectionId, char* key);
void frankenphp_ws_deleteStoredInformation(char* connectionId, char* key);
void frankenphp_ws_clearStoredInformation(char* connectionId);
int frankenphp_ws_hasStoredInformation(char* connectionId, char* key);
void frankenphp_ws_listStoredInformationKeys(void* array, char* connectionId);

// Tag expression logic functions
void frankenphp_ws_sendToTagExpression(char* expression, char* data, int data_len, char* route);
void frankenphp_ws_getClientsByTagExpression(void* array, char* expression);

// Connection management functions
int frankenphp_ws_renameConnection(char* currentId, char* newId);

#endif