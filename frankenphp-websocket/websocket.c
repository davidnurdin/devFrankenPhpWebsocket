#include <php.h>
#include "websocket.h"
#include "websocket_arginfo.h"

// Contient les symboles exportés par Go
#include "_cgo_export.h"

// Holds the current return_value while building the array
static zval* frankenphp_ws_current_retval = NULL;

void frankenphp_ws_addClient(const char* id)
{
    if (frankenphp_ws_current_retval == NULL) {
        return;
    }
    add_next_index_string(frankenphp_ws_current_retval, id);
}

PHP_FUNCTION(frankenphp_ws_getClients)
{
    ZEND_PARSE_PARAMETERS_NONE();
    array_init(return_value);
    frankenphp_ws_current_retval = return_value;
    // Go will iterate clients and call back frankenphp_ws_addClient for each
    frankenphp_ws_getClients();
    frankenphp_ws_current_retval = NULL;
}

PHP_FUNCTION(frankenphp_ws_send)
{
    char *connectionId = NULL;
    size_t connectionId_len = 0;
    char *data = NULL;
    size_t data_len = 0;

    ZEND_PARSE_PARAMETERS_START(2, 2)
        Z_PARAM_STRING(connectionId, connectionId_len)
        Z_PARAM_STRING(data, data_len)
    ZEND_PARSE_PARAMETERS_END();

    frankenphp_ws_send(connectionId, data, (int)data_len);
}


zend_module_entry ext_module_entry = {
    STANDARD_MODULE_HEADER,
    "frankenphp_websocket",
    ext_functions, /* Functions */
    NULL,          /* MINIT */
    NULL,          /* MSHUTDOWN */
    NULL,          /* RINIT */
    NULL,          /* RSHUTDOWN */
    NULL,          /* MINFO */
    "0.1.1",
    STANDARD_MODULE_PROPERTIES
};
