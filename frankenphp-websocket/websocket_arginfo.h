/* This is a generated file, edit the .stub.php file instead.
 * Stub hash: 1f8a1de4642a4424394247206ce0b59f635e79da */

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClients, 0, 0, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, route, IS_STRING, 1, "null")
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_send, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, route, IS_STRING, 1, "null")
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_tagClient, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_untagClient arginfo_frankenphp_ws_tagClient

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_clearTagClient, 0, 1, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getTags, 0, 0, IS_ARRAY, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClientsByTag, 0, 1, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_sendToTag, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, tag, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, route, IS_STRING, 1, "null")
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_setStoredInformation, 0, 3, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, value, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getStoredInformation, 0, 2, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_deleteStoredInformation, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_clearStoredInformation arginfo_frankenphp_ws_clearTagClient

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_hasStoredInformation, 0, 2, _IS_BOOL, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_listStoredInformationKeys, 0, 1, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_sendToTagExpression, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, expression, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, route, IS_STRING, 1, "null")
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClientsByTagExpression, 0, 1, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO(0, expression, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_listRoutes arginfo_frankenphp_ws_getTags

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_renameConnection, 0, 2, _IS_BOOL, 0)
	ZEND_ARG_TYPE_INFO(0, currentId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, newId, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_global_set, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, value, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, expireSeconds, IS_LONG, 0, "0")
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_global_get, 0, 1, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_global_has, 0, 1, _IS_BOOL, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
ZEND_END_ARG_INFO()

#define arginfo_frankenphp_ws_global_delete arginfo_frankenphp_ws_global_has

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_searchStoredInformation, 0, 3, IS_ARRAY, 0)
	ZEND_ARG_TYPE_INFO(0, key, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, op, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, value, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, route, IS_STRING, 1, "null")
ZEND_END_ARG_INFO()

ZEND_FUNCTION(frankenphp_ws_getClients);
ZEND_FUNCTION(frankenphp_ws_send);
ZEND_FUNCTION(frankenphp_ws_tagClient);
ZEND_FUNCTION(frankenphp_ws_untagClient);
ZEND_FUNCTION(frankenphp_ws_clearTagClient);
ZEND_FUNCTION(frankenphp_ws_getTags);
ZEND_FUNCTION(frankenphp_ws_getClientsByTag);
ZEND_FUNCTION(frankenphp_ws_sendToTag);
ZEND_FUNCTION(frankenphp_ws_setStoredInformation);
ZEND_FUNCTION(frankenphp_ws_getStoredInformation);
ZEND_FUNCTION(frankenphp_ws_deleteStoredInformation);
ZEND_FUNCTION(frankenphp_ws_clearStoredInformation);
ZEND_FUNCTION(frankenphp_ws_hasStoredInformation);
ZEND_FUNCTION(frankenphp_ws_listStoredInformationKeys);
ZEND_FUNCTION(frankenphp_ws_sendToTagExpression);
ZEND_FUNCTION(frankenphp_ws_getClientsByTagExpression);
ZEND_FUNCTION(frankenphp_ws_listRoutes);
ZEND_FUNCTION(frankenphp_ws_renameConnection);
ZEND_FUNCTION(frankenphp_ws_global_set);
ZEND_FUNCTION(frankenphp_ws_global_get);
ZEND_FUNCTION(frankenphp_ws_global_has);
ZEND_FUNCTION(frankenphp_ws_global_delete);
ZEND_FUNCTION(frankenphp_ws_searchStoredInformation);

static const zend_function_entry ext_functions[] = {
	ZEND_FE(frankenphp_ws_getClients, arginfo_frankenphp_ws_getClients)
	ZEND_FE(frankenphp_ws_send, arginfo_frankenphp_ws_send)
	ZEND_FE(frankenphp_ws_tagClient, arginfo_frankenphp_ws_tagClient)
	ZEND_FE(frankenphp_ws_untagClient, arginfo_frankenphp_ws_untagClient)
	ZEND_FE(frankenphp_ws_clearTagClient, arginfo_frankenphp_ws_clearTagClient)
	ZEND_FE(frankenphp_ws_getTags, arginfo_frankenphp_ws_getTags)
	ZEND_FE(frankenphp_ws_getClientsByTag, arginfo_frankenphp_ws_getClientsByTag)
	ZEND_FE(frankenphp_ws_sendToTag, arginfo_frankenphp_ws_sendToTag)
	ZEND_FE(frankenphp_ws_setStoredInformation, arginfo_frankenphp_ws_setStoredInformation)
	ZEND_FE(frankenphp_ws_getStoredInformation, arginfo_frankenphp_ws_getStoredInformation)
	ZEND_FE(frankenphp_ws_deleteStoredInformation, arginfo_frankenphp_ws_deleteStoredInformation)
	ZEND_FE(frankenphp_ws_clearStoredInformation, arginfo_frankenphp_ws_clearStoredInformation)
	ZEND_FE(frankenphp_ws_hasStoredInformation, arginfo_frankenphp_ws_hasStoredInformation)
	ZEND_FE(frankenphp_ws_listStoredInformationKeys, arginfo_frankenphp_ws_listStoredInformationKeys)
	ZEND_FE(frankenphp_ws_sendToTagExpression, arginfo_frankenphp_ws_sendToTagExpression)
	ZEND_FE(frankenphp_ws_getClientsByTagExpression, arginfo_frankenphp_ws_getClientsByTagExpression)
	ZEND_FE(frankenphp_ws_listRoutes, arginfo_frankenphp_ws_listRoutes)
	ZEND_FE(frankenphp_ws_renameConnection, arginfo_frankenphp_ws_renameConnection)
	ZEND_FE(frankenphp_ws_global_set, arginfo_frankenphp_ws_global_set)
	ZEND_FE(frankenphp_ws_global_get, arginfo_frankenphp_ws_global_get)
	ZEND_FE(frankenphp_ws_global_has, arginfo_frankenphp_ws_global_has)
	ZEND_FE(frankenphp_ws_global_delete, arginfo_frankenphp_ws_global_delete)
	ZEND_FE(frankenphp_ws_searchStoredInformation, arginfo_frankenphp_ws_searchStoredInformation)
	ZEND_FE_END
};
