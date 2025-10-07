<?php

foreach (frankenphp_ws_getClients() as $client)
{
    frankenphp_ws_send($client,"HELLO " . rand(1,1000));
}

// var_dump(frankenphp_ws_getClients());
