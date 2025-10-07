/* This is a generated file, edit the .stub.php file instead.
 * Stub hash: 25fbd17a9c8ffd2e2d44650d43c0ff8913b942ed */

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_getClients, 0, 0, IS_ARRAY, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_frankenphp_ws_send, 0, 2, IS_VOID, 0)
	ZEND_ARG_TYPE_INFO(0, connectionId, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO(0, data, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_FUNCTION(frankenphp_ws_getClients);
ZEND_FUNCTION(frankenphp_ws_send);

static const zend_function_entry ext_functions[] = {
	ZEND_FE(frankenphp_ws_getClients, arginfo_frankenphp_ws_getClients)
	ZEND_FE(frankenphp_ws_send, arginfo_frankenphp_ws_send)
	ZEND_FE_END
};
