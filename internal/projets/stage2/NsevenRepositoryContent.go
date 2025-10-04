package stage2

import "fmt"

func NsevenMongoRepositoryContent(moduleName string) string {
	return fmt.Sprintf(`package nsevenrepository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"%s/internal/domain/nseven"
)

type mongoNsevenRepository struct {
	collection *mongo.Collection
}

// NewMongoNsevenRepository crée une nouvelle instance du repository MongoDB pour Nseven
func NewMongoNsevenRepository(database *mongo.Database) nseven.NsevenRepository {
	return &mongoNsevenRepository{
		collection: database.Collection("nsevens"),
	}
}

// FindByID récupère un Nseven par son ID
func (r *mongoNsevenRepository) FindByID(ctx context.Context, id string) (*nseven.Nseven, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("ID invalide: %%w", err)
	}

	var result nseven.Nseven
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("nseven non trouvé")
		}
		return nil, fmt.Errorf("erreur lors de la récupération: %%w", err)
	}

	return &result, nil
}

// FindAll récupère tous les Nseven
func (r *mongoNsevenRepository) FindAll(ctx context.Context) ([]*nseven.Nseven, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération: %%w", err)
	}
	defer cursor.Close(ctx)

	var results []*nseven.Nseven
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage: %%w", err)
	}

	return results, nil
}

// Create crée un nouveau Nseven
func (r *mongoNsevenRepository) Create(ctx context.Context, nsevenEntity *nseven.Nseven) error {
	result, err := r.collection.InsertOne(ctx, nsevenEntity)
	if err != nil {
		return fmt.Errorf("erreur lors de la création: %%w", err)
	}

	// Mettre à jour l'ID de l'entité avec celui généré par MongoDB
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		nsevenEntity.ID = oid.Hex()
	}

	return nil
}

// Update met à jour un Nseven existant
func (r *mongoNsevenRepository) Update(ctx context.Context, nsevenEntity *nseven.Nseven) error {
	objectID, err := primitive.ObjectIDFromHex(nsevenEntity.ID)
	if err != nil {
		return fmt.Errorf("ID invalide: %%w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"message": nsevenEntity.Message,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour: %%w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("nseven non trouvé")
	}

	return nil
}

// Delete supprime un Nseven par son ID
func (r *mongoNsevenRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("ID invalide: %%w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression: %%w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("nseven non trouvé")
	}

	return nil
}
`, moduleName)
}
