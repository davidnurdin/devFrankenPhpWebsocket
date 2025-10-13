<?php

echo "Get all WS clients ... <br>" ;

var_dump(frankenphp_ws_listRoutes());

foreach (frankenphp_ws_getClients('/route2') as $client)
{
    echo "Send HELLO to " . $client . "<br>" ;
    frankenphp_ws_send($client,"HELLO " . rand(1,1000));
}

var_dump(frankenphp_ws_getTags());

// var_dump(frankenphp_ws_getClients());
frankenphp_ws_sendToTag('group_2','group 2 TAG');

