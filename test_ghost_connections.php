<?php

/**
 * Test des connexions fantômes WebSocket
 * 
 * Ce script démontre l'utilisation des connexions fantômes :
 * - activateGhost($clientId) : Marque une connexion comme fantôme
 * - releaseGhost($clientId) : Libère une connexion fantôme et déclenche les événements
 * - isGhost($clientId) : Vérifie si une connexion est fantôme
 * 
 * Flux d'événements :
 * 1. Connexion normale : open -> message -> beforeClose -> close
 * 2. Connexion fantôme : open -> message -> (disconnect ignoré)
 * 3. Libération fantôme : ghostConnectionClose -> beforeClose -> close
 */

echo "=== Test des connexions fantômes WebSocket ===\n\n";

// Simuler une connexion
$clientId = "test-client-" . uniqid();
echo "1. Connexion simulée : $clientId\n";

// Vérifier l'état initial (pas fantôme)
$isGhost = frankenphp_ws_isGhost($clientId);
echo "2. État initial (fantôme) : " . ($isGhost ? "OUI" : "NON") . "\n";

// Activer le mode fantôme
echo "3. Activation du mode fantôme...\n";
$activated = frankenphp_ws_activateGhost($clientId);
echo "   Résultat : " . ($activated ? "SUCCÈS" : "ÉCHEC") . "\n";

// Vérifier l'état fantôme
$isGhost = frankenphp_ws_isGhost($clientId);
echo "4. État après activation (fantôme) : " . ($isGhost ? "OUI" : "NON") . "\n";

// Simuler une déconnexion (normalement ignorée pour les fantômes)
echo "5. Simulation d'une déconnexion...\n";
echo "   → Pour une connexion fantôme, les événements beforeClose/close sont ignorés\n";
echo "   → La connexion reste 'vivante' côté serveur\n";

// Libérer la connexion fantôme
echo "6. Libération de la connexion fantôme...\n";
$released = frankenphp_ws_releaseGhost($clientId);
echo "   Résultat : " . ($released ? "SUCCÈS" : "ÉCHEC") . "\n";
echo "   → Événements déclenchés : ghostConnectionClose -> beforeClose -> close\n";

// Vérifier l'état final
$isGhost = frankenphp_ws_isGhost($clientId);
echo "7. État final (fantôme) : " . ($isGhost ? "OUI" : "NON") . "\n";

echo "\n=== Cas d'usage typiques ===\n\n";

echo "1. GESTION DE RECONNEXION :\n";
echo "   - Client se déconnecte temporairement\n";
echo "   - activateGhost() pour éviter le nettoyage\n";
echo "   - Client se reconnecte avec le même ID\n";
echo "   - releaseGhost() pour finaliser la déconnexion précédente\n\n";

echo "2. MAINTENANCE PROGRAMMÉE :\n";
echo "   - activateGhost() avant maintenance\n";
echo "   - Les déconnexions sont ignorées\n";
echo "   - releaseGhost() après maintenance pour nettoyer\n\n";

echo "3. GESTION D'ÉTAT TEMPORAIRE :\n";
echo "   - activateGhost() pendant une opération critique\n";
echo "   - Évite les déconnexions accidentelles\n";
echo "   - releaseGhost() une fois l'opération terminée\n\n";

echo "=== API des connexions fantômes ===\n\n";

echo "fonctions disponibles :\n";
echo "- frankenphp_ws_activateGhost(string \$connectionId): bool\n";
echo "  → Marque une connexion comme fantôme\n";
echo "  → Les déconnexions sont ignorées\n\n";

echo "- frankenphp_ws_releaseGhost(string \$connectionId): bool\n";
echo "  → Libère une connexion fantôme\n";
echo "  → Déclenche : ghostConnectionClose -> beforeClose -> close\n\n";

echo "- frankenphp_ws_isGhost(string \$connectionId): bool\n";
echo "  → Vérifie si une connexion est en mode fantôme\n\n";

echo "événements :\n";
echo "- ghostConnectionClose : Déclenché avant beforeClose lors de releaseGhost()\n";
echo "- beforeClose : Déclenché normalement ou lors de releaseGhost()\n";
echo "- close : Déclenché normalement ou lors de releaseGhost()\n\n";

echo "=== Test terminé ===\n";
