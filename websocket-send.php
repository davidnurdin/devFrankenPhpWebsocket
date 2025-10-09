<?php

echo "Get all WS clients ... <br>" ;


foreach (frankenphp_ws_getClients() as $client)
{
    echo "Send HELLO to " . $client . "<br>" ;
    frankenphp_ws_send($client,"HELLO " . rand(1,1000));
}

// var_dump(frankenphp_ws_getClients());
