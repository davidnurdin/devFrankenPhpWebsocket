<?php

// Require the Composer autoloader here if needed (API Platform, Symfony, etc.)
//require __DIR__ . '/vendor/autoload.php';

// Handler outside the loop for better performance (doing less work)
$handler = static function (array $event): array  {
    // $event['Type']        => 'open' | 'message' | 'close'
    // $event['Connection']  => identifiant de connexion (string)
    // $event['Payload']     => données associées

    // TODO voir => bug : frankenphp_ws_getClients crash le serveur en cas d'open/close simultané

    file_put_contents('php://stderr','Result of frankenphp_ws_getClients : ' . var_export(frankenphp_ws_getClients(),true));


    $group = rand(1,5);
    frankenphp_ws_tagClient($event['Connection'],'group_' . $group);
    
    file_put_contents('php://stderr', "Handler called with " . var_export($event, true) . "\n");

    if (($event['Type'] ?? null) === 'message') {
        foreach (frankenphp_ws_getClients() as $client)
        {
            frankenphp_ws_send($client,"MSG FROM ANOTHER WS => " . (string)($event['Payload']));
        }

        
        return ['message' => 'Hello from PHP : ' . (string)($event['Payload'] ?? '')];
    }

    // Pour open/close, aucune réponse particulière n'est requise
    return ['ok' => true];
};

file_put_contents('php://stderr', "WebSocket worker started with PID " . getmypid() . "\n");


$maxRequests = (int)($_SERVER['MAX_REQUESTS'] ?? 0); // illimité si 0

for ($nbRequests = 0; !$maxRequests || $nbRequests < $maxRequests; ++$nbRequests) {
    file_put_contents('php://stderr', "Handled request #" . ($nbRequests + 1) . "\n");
    $keepRunning = \frankenphp_handle_request($handler);
    file_put_contents('php://stderr', "Request handled, keepRunning=" . ($keepRunning ? 'true' : 'false') . "\n");

    // Call the garbage collector to reduce the chances of it being triggered in the middle of the handling of a request
    gc_collect_cycles();

    if (!$keepRunning) {
      break;
    }
}

file_put_contents('php://stderr', "Max requests reached, exiting\n");