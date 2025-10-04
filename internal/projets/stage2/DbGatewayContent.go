package stage2

func DbGatewayContent() string {
	return `package dbgateway

import "context"

// Database représente le gateway pour l'accès à la base de données
type Database interface {
	// Connect établit la connexion à la base de données
	Connect(ctx context.Context) error

	// Disconnect ferme la connexion à la base de données
	Disconnect(ctx context.Context) error

	// Ping vérifie la connexion à la base de données
	Ping(ctx context.Context) error

	// GetClient retourne le client de base de données natif
	GetClient() interface{}
}
`
}
