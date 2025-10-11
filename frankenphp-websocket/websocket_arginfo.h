/* This is a generated file, edit the .stub.php file instead.
 * Stub hash: 5eea73fce8bd679f070ce57e3a1b3d667afe1a0e */

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClients, 0, 0, IS_ARRAY, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_send, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_tagClient, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_untagClient arginfo_frankenphp_ws_tagClient

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_clearTagClient, 0, 1, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_getTags arginfo_frankenphp_ws_getClients

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClientsByTag, 0, 1, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_sendToTag, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_FUNCTION(frankenphp_ws_getClients);
ZEND_FUNCTION(frankenphp_ws_send);
ZEND_FUNCTION(frankenphp_ws_tagClient);
ZEND_FUNCTION(frankenphp_ws_untagClient);
ZEND_FUNCTION(frankenphp_ws_clearTagClient);
ZEND_FUNCTION(frankenphp_ws_getTags);
ZEND_FUNCTION(frankenphp_ws_getClientsByTag);
ZEND_FUNCTION(frankenphp_ws_sendToTag);

static const zend_function_entry ext_functions[] = {
	ZEND_FE(frankenphp_ws_getClients, arginfo_frankenphp_ws_getClients)
	ZEND_FE(frankenphp_ws_send, arginfo_frankenphp_ws_send)
	ZEND_FE(frankenphp_ws_tagClient, arginfo_frankenphp_ws_tagClient)
	ZEND_FE(frankenphp_ws_untagClient, arginfo_frankenphp_ws_untagClient)
	ZEND_FE(frankenphp_ws_clearTagClient, arginfo_frankenphp_ws_clearTagClient)
	ZEND_FE(frankenphp_ws_getTags, arginfo_frankenphp_ws_getTags)
	ZEND_FE(frankenphp_ws_getClientsByTag, arginfo_frankenphp_ws_getClientsByTag)
	ZEND_FE(frankenphp_ws_sendToTag, arginfo_frankenphp_ws_sendToTag)
	ZEND_FE_END
};
