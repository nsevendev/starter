package stage2

func NsevenEntityContent() string {
	return `package nseven

// Nseven représente l'entité métier Nseven
type Nseven struct {
	ID      string ` + "`bson:\"_id,omitempty\" json:\"id\"`" + `
	Message string ` + "`bson:\"message\" json:\"message\"`" + `
}

// NewNseven crée une nouvelle instance de Nseven
func NewNseven(message string) *Nseven {
	return &Nseven{
		Message: message,
	}
}

// GetGreeting retourne le message de bienvenue
func (n *Nseven) GetGreeting() string {
	return n.Message
}
`
}

func NsevenRepositoryInterfaceContent() string {
	return `package nseven

import "context"

// NsevenRepository définit les opérations de persistance pour l'entité Nseven
type NsevenRepository interface {
	// FindByID récupère un Nseven par son ID
	FindByID(ctx context.Context, id string) (*Nseven, error)

	// FindAll récupère tous les Nseven
	FindAll(ctx context.Context) ([]*Nseven, error)

	// Create crée un nouveau Nseven
	Create(ctx context.Context, nseven *Nseven) error

	// Update met à jour un Nseven existant
	Update(ctx context.Context, nseven *Nseven) error

	// Delete supprime un Nseven par son ID
	Delete(ctx context.Context, id string) error
}
`
}
