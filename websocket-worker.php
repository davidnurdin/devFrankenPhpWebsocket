<?php

// Require the Composer autoloader here if needed (API Platform, Symfony, etc.)
//require __DIR__ . '/vendor/autoload.php';

// Handler outside the loop for better performance (doing less work)
$handler = static function (string $msg): array  {
	// Do something with the gRPC request
    return ['message' => "Hello from PHP : " . $msg];
};

file_put_contents('php://stderr', "WebSocket worker started with PID " . getmypid() . "\n");


$maxRequests = (int)($_SERVER['MAX_REQUESTS'] ?? 0);
$maxRequests = 2 ;

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