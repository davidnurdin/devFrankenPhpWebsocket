#include <php.h>
#include "websocket.h"
#include "websocket_arginfo.h"

// Contient les symboles exportés par Go
#include "_cgo_export.h"

void frankenphp_ws_addClient(zval* array, const char* id)
{
    if (array == NULL) {
        return;
    }
    add_next_index_string(array, id);
}

PHP_FUNCTION(frankenphp_ws_getClients)
{
    char *route = NULL;
    size_t route_len = 0;

    ZEND_PARSE_PARAMETERS_START(0, 1)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(route, route_len)
    ZEND_PARSE_PARAMETERS_END();

    array_init(return_value);
    // Go will iterate clients and call back frankenphp_ws_addClient for each
    frankenphp_ws_getClients((void*)return_value, route);
}

PHP_FUNCTION(frankenphp_ws_send)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *data = NULL;
    size_t data_len = 0;
    char *route = NULL;
    size_t route_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(data, data_len)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(route, route_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_send(connectionId, data, (int)data_len, route);
}

PHP_FUNCTION(frankenphp_ws_tagClient)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *tag = NULL;
    size_t tag_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(tag, tag_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_tagClient(connectionId, tag);
}

PHP_FUNCTION(frankenphp_ws_untagClient)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *tag = NULL;
    size_t tag_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(tag, tag_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_untagClient(connectionId, tag);
}

PHP_FUNCTION(frankenphp_ws_clearTagClient)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(connectionId, connectionId_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_clearTagClient(connectionId);
}

PHP_FUNCTION(frankenphp_ws_getTags)
{
    ZEND_PARSE_PARAMETERS_NONE();
    array_init(return_value);
    // Go will iterate tags and call back frankenphp_ws_addClient for each
    frankenphp_ws_getTags((void*)return_value);
}

PHP_FUNCTION(frankenphp_ws_getClientsByTag)
{
    char *tag = NULL;
    size_t tag_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(tag, tag_len)
    ZEND_PARSE_PARAMETERS_END();

    array_init(return_value);
    // Go will iterate clients and call back frankenphp_ws_addClient for each
    frankenphp_ws_getClientsByTag((void*)return_value, tag);
}

PHP_FUNCTION(frankenphp_ws_sendToTag)
{
    char *tag = NULL;
    size_t tag_len = 0;
    char *data = NULL;
    size_t data_len = 0;
    char *route = NULL;
    size_t route_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_STRING(tag, tag_len)
        Z_PARAM_STRING(data, data_len)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(route, route_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_sendToTag(tag, data, (int)data_len, route);
}

PHP_FUNCTION(frankenphp_ws_setStoredInformation)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *key = NULL;
    size_t key_len = 0;
    char *value = NULL;
    size_t value_len = 0;

    ZEND_PARSE_PARAMETERS_START(3, 3)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(key, key_len)
        Z_PARAM_STRING(value, value_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_setStoredInformation(connectionId, key, value);
}

PHP_FUNCTION(frankenphp_ws_getStoredInformation)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *key = NULL;
    size_t key_len = 0;
    char *result = NULL;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    result = frankenphp_ws_getStoredInformation(connectionId, key);
    if (result != NULL) {
        RETURN_STRING(result);
    } else {
        RETURN_EMPTY_STRING();
    }
}

PHP_FUNCTION(frankenphp_ws_deleteStoredInformation)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *key = NULL;
    size_t key_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_deleteStoredInformation(connectionId, key);
}

PHP_FUNCTION(frankenphp_ws_clearStoredInformation)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(connectionId, connectionId_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_clearStoredInformation(connectionId);
}

PHP_FUNCTION(frankenphp_ws_hasStoredInformation)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *key = NULL;
    size_t key_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    RETURN_BOOL(frankenphp_ws_hasStoredInformation(connectionId, key) == 1);
}

PHP_FUNCTION(frankenphp_ws_listStoredInformationKeys)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(connectionId, connectionId_len)
    ZEND_PARSE_PARAMETERS_END();

    array_init(return_value);
    // Go will iterate keys and call back frankenphp_ws_addClient for each
    frankenphp_ws_listStoredInformationKeys((void*)return_value, connectionId);
}

PHP_FUNCTION(frankenphp_ws_sendToTagExpression)
{
    char *expression = NULL;
    size_t expression_len = 0;
    char *data = NULL;
    size_t data_len = 0;
    char *route = NULL;
    size_t route_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_STRING(expression, expression_len)
        Z_PARAM_STRING(data, data_len)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(route, route_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_sendToTagExpression(expression, data, (int)data_len, route);
}

PHP_FUNCTION(frankenphp_ws_getClientsByTagExpression)
{
    char *expression = NULL;
    size_t expression_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(expression, expression_len)
    ZEND_PARSE_PARAMETERS_END();

    array_init(return_value);
    // Go will iterate clients and call back frankenphp_ws_addClient for each
    frankenphp_ws_getClientsByTagExpression((void*)return_value, expression);
}

PHP_FUNCTION(frankenphp_ws_listRoutes)
{
    ZEND_PARSE_PARAMETERS_NONE();
    array_init(return_value);
    // Go will iterate routes and call back frankenphp_ws_addClient for each
    frankenphp_ws_listRoutes((void*)return_value);
}

PHP_FUNCTION(frankenphp_ws_renameConnection)
{
    char *currentId = NULL;
    size_t currentId_len = 0;
    char *newId = NULL;
    size_t newId_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(currentId, currentId_len)
        Z_PARAM_STRING(newId, newId_len)
    ZEND_PARSE_PARAMETERS_END();

    ZVAL_BOOL(return_value, frankenphp_ws_renameConnection(currentId, newId));
}

PHP_FUNCTION(frankenphp_ws_searchStoredInformation)
{
    char *key = NULL; size_t key_len = 0;
    char *op = NULL; size_t op_len = 0;
    char *value = NULL; size_t value_len = 0;
    char *route = NULL; size_t route_len = 0;

    ZEND_PARSE_PARAMETERS_START(3, 4)
        Z_PARAM_STRING(key, key_len)
        Z_PARAM_STRING(op, op_len)
        Z_PARAM_STRING(value, value_len)
        Z_PARAM_OPTIONAL
        Z_PARAM_STRING(route, route_len)
    ZEND_PARSE_PARAMETERS_END();

    array_init(return_value);
    frankenphp_ws_searchStoredInformation((void*)return_value, key, op, value, route);
}

// ===== Global info glue =====
PHP_FUNCTION(frankenphp_ws_global_set)
{
    char *key = NULL; size_t key_len = 0;
    char *value = NULL; size_t value_len = 0;
    zend_long expireSeconds = 0;

    ZEND_PARSE_PARAMETERS_START(2, 3)
        Z_PARAM_STRING(key, key_len)
        Z_PARAM_STRING(value, value_len)
        Z_PARAM_OPTIONAL
        Z_PARAM_LONG(expireSeconds)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_global_set(key, value, (int)expireSeconds);
}

PHP_FUNCTION(frankenphp_ws_global_get)
{
    char *key = NULL; size_t key_len = 0; char *result = NULL;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    result = frankenphp_ws_global_get(key);
    if (result != NULL) {
        RETURN_STRING(result);
    } else {
        RETURN_EMPTY_STRING();
    }
}

PHP_FUNCTION(frankenphp_ws_global_has)
{
    char *key = NULL; size_t key_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    RETURN_BOOL(frankenphp_ws_global_has(key) == 1);
}

PHP_FUNCTION(frankenphp_ws_global_delete)
{
    char *key = NULL; size_t key_len = 0;

    ZEND_PARSE_PARAMETERS_START(1, 1)
        Z_PARAM_STRING(key, key_len)
    ZEND_PARSE_PARAMETERS_END();

    RETURN_BOOL(frankenphp_ws_global_delete(key) == 1);
}

static PHP_MINIT_FUNCTION(frankenphp_websocket)
{
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_EQ", "eq", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_NEQ", "neq", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_PREFIX", "prefix", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_SUFFIX", "suffix", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_CONTAINS", "contains", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_IEQ", "ieq", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_IPREFIX", "iprefix", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_ISUFFIX", "isuffix", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_ICONTAINS", "icontains", CONST_CS | CONST_PERSISTENT);
    REGISTER_STRING_CONSTANT("FRANKENPHP_WS_OP_REGEX", "regex", CONST_CS | CONST_PERSISTENT);
    return SUCCESS;
}

zend_module_entry ext_module_entry = {
    STANDARD_MODULE_HEADER,
    "frankenphp_websocket",
    ext_functions, /* Functions */
    PHP_MINIT(frankenphp_websocket),          /* MINIT */
    NULL,          /* MSHUTDOWN */
    NULL,          /* RINIT */
    NULL,          /* RSHUTDOWN */
    NULL,          /* MINFO */
    "0.1.1",
    STANDARD_MODULE_PROPERTIES
};
