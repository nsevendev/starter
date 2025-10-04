package stage2

import "fmt"

func NsevenControllerContent(moduleName string) string {
	return fmt.Sprintf(`package nsevencontroller

import (
	"%s/internal/application/gateway/httpgateway"
	"%s/internal/application/usecase/nsevenusecase"
)

type NsevenController struct {
	useCase   *nsevenusecase.NsevenUseCase
	prefixUrl string
}

// New crée une nouvelle instance du controller Nseven
func New(useCase *nsevenusecase.NsevenUseCase) *NsevenController {
	return &NsevenController{
		useCase:   useCase,
		prefixUrl: "/nseven",
	}
}

// RegisterRoutes enregistre les routes du controller
func (c *NsevenController) RegisterRoutes(r httpgateway.Router) {
	r.Handle("POST", "/nseven", c.CreateNseven)
	r.Handle("GET", "/nseven", c.GetAllNseven)
}
`, moduleName, moduleName)
}

func NsevenCreateContent(moduleName string) string {
	return fmt.Sprintf(`package nsevencontroller

import "%s/internal/application/gateway/httpgateway"

// CreateNseven crée un nouveau Nseven
// @Summary Créer un Nseven
// @Description Crée un nouveau Nseven avec le message "Bonjour Nseven"
// @Tags Nseven
// @Accept json
// @Produce json
// @Success 201 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nseven [post]
func (c *NsevenController) CreateNseven(ctx httpgateway.Context) {
	nsevenEntity, err := c.useCase.CreateNseven(ctx.Request().Context())
	if err != nil {
		ctx.InternalServerError("Erreur lors de la création", err.Error())
		return
	}

	ctx.Created("Le nseven à été créé", nsevenEntity)
}
`, moduleName)
}

func NsevenGetAllContent(moduleName string) string {
	return fmt.Sprintf(`package nsevencontroller

import "%s/internal/application/gateway/httpgateway"

// GetAllNseven récupère tous les Nseven
// @Summary Récupérer tous les Nseven
// @Description Récupère la liste de tous les Nseven
// @Tags Nseven
// @Accept json
// @Produce json
// @Success 200 {array} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nseven [get]
func (c *NsevenController) GetAllNseven(ctx httpgateway.Context) {
	nsevens, err := c.useCase.GetAll(ctx.Request().Context())
	if err != nil {
		ctx.InternalServerError("Erreur lors de la récupération", err.Error())
		return
	}

	ctx.Success("Récupération de tous les nseven", nsevens)
}
`, moduleName)
}
