package stage2

import "fmt"

func NsevenUseCaseContent(moduleName string) string {
	return fmt.Sprintf(`package nsevenusecase

import (
	"context"
	"%s/internal/domain/nseven"
)

// NsevenUseCase gère la logique métier pour les Nseven
type NsevenUseCase struct {
	repo nseven.NsevenRepository
}

// NewNsevenUseCase crée une nouvelle instance du use case
func NewNsevenUseCase(repo nseven.NsevenRepository) *NsevenUseCase {
	return &NsevenUseCase{
		repo: repo,
	}
}

// GetByID récupère un Nseven par son ID
func (uc *NsevenUseCase) GetByID(ctx context.Context, id string) (*nseven.Nseven, error) {
	return uc.repo.FindByID(ctx, id)
}

// GetAll récupère tous les Nseven
func (uc *NsevenUseCase) GetAll(ctx context.Context) ([]*nseven.Nseven, error) {
	return uc.repo.FindAll(ctx)
}

// CreateNseven crée un nouveau Nseven avec le message "Bonjour Nseven"
func (uc *NsevenUseCase) CreateNseven(ctx context.Context) (*nseven.Nseven, error) {
	nsevenEntity := nseven.NewNseven("Bonjour Nseven")

	err := uc.repo.Create(ctx, nsevenEntity)
	if err != nil {
		return nil, err
	}

	return nsevenEntity, nil
}

// UpdateMessage met à jour le message d'un Nseven
func (uc *NsevenUseCase) UpdateMessage(ctx context.Context, id string, newMessage string) (*nseven.Nseven, error) {
	nsevenEntity, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	nsevenEntity.Message = newMessage

	err = uc.repo.Update(ctx, nsevenEntity)
	if err != nil {
		return nil, err
	}

	return nsevenEntity, nil
}

// Delete supprime un Nseven
func (uc *NsevenUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
`, moduleName)
}
