package stage2

import "fmt"

func MongoAdapterContent(moduleName string) string {
	return fmt.Sprintf(`package mongoadapter

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"%s/internal/application/gateway/dbgateway"
	"%s/internal/application/gateway/loggateway"
	"time"
)

type mongoAdapter struct {
	client   *mongo.Client
	database *mongo.Database
	uri      string
	dbName   string
	logger   loggateway.Logger
}

// New crée une nouvelle instance de l'adaptateur MongoDB
func New(uri, dbName string, logger loggateway.Logger) dbgateway.Database {
	return &mongoAdapter{
		uri:    uri,
		dbName: dbName,
		logger: logger,
	}
}

// Connect établit la connexion à MongoDB
func (m *mongoAdapter) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(m.uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.logger.Ef("Erreur lors de la connexion à MongoDB: %%v", err)
		return fmt.Errorf("erreur de connexion MongoDB: %%w", err)
	}

	// Vérifier la connexion
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		m.logger.Ef("Erreur lors du ping MongoDB: %%v", err)
		return fmt.Errorf("erreur de ping MongoDB: %%w", err)
	}

	m.client = client
	m.database = client.Database(m.dbName)

	m.logger.If("Connexion à MongoDB établie avec succès - Base: %%s", m.dbName)

	return nil
}

// Disconnect ferme la connexion à MongoDB
func (m *mongoAdapter) Disconnect(ctx context.Context) error {
	if m.client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		m.logger.Ef("Erreur lors de la déconnexion de MongoDB: %%v", err)
		return fmt.Errorf("erreur de déconnexion MongoDB: %%w", err)
	}

	m.logger.If("Déconnexion de MongoDB réussie")
	return nil
}

// Ping vérifie la connexion à MongoDB
func (m *mongoAdapter) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("client MongoDB non initialisé")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := m.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("erreur de ping MongoDB: %%w", err)
	}

	return nil
}

// GetClient retourne le client MongoDB natif
func (m *mongoAdapter) GetClient() interface{} {
	return m.client
}

// GetDatabase retourne l'instance de la base de données MongoDB
func (m *mongoAdapter) GetDatabase() *mongo.Database {
	return m.database
}

// GetCollection retourne une collection MongoDB spécifique
func (m *mongoAdapter) GetCollection(name string) *mongo.Collection {
	return m.database.Collection(name)
}
`, moduleName, moduleName)
}
