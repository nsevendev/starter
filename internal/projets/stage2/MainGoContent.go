package stage2

import "fmt"

func MainGoContent(moduleName string) string {
	return fmt.Sprintf(`package main

import (
	"context"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nsevenpack/env/env"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"%s/docs"
	"%s/internal/application/controller/nsevencontroller"
	"%s/internal/application/controller/testcontroller"
	"%s/internal/application/gateway/dbgateway"
	"%s/internal/application/gateway/httpgateway"
	"%s/internal/application/gateway/loggateway"
	"%s/internal/application/usecase/nsevenusecase"
	"%s/internal/infrastructure/adapter/ginadapter"
	"%s/internal/infrastructure/adapter/loggeradapter"
	"%s/internal/infrastructure/adapter/mongoadapter"
	"%s/internal/infrastructure/repository/nsevenrepository"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	loggerAdapter loggateway.Logger
	dbAdapter     dbgateway.Database
)

// @title nseven api
// @version 1.0
// @description API service nseven api
// @schemes https
// @securityDefinitions.apikey BearerAuth
// @in headers
// @name Authorization
func main() {
	ctx := context.Background()

	initDatabase(ctx)
	defer closeDatabase(ctx)

	s := gin.Default()
	host := "0.0.0.0"
	port := env.Get("PORT")
	hostTraefikApi := extractStringInBacktick(env.Get("HOST_TRAEFIK_API"))

	infoServer(hostTraefikApi)
	setSwaggerOpt(hostTraefikApi)
	setCors(s)
	router(s)

	startServerWithGracefulShutdown(s, host, port)
}

func init() {
	appEnv := env.Get("APP_ENV")
	loggerAdapter = loggeradapter.New(appEnv)

	dbUri := env.Get("DB_URI")
	dbName := env.Get("DB_NAME")
	dbAdapter = mongoadapter.New(dbUri, dbName, loggerAdapter)
}

func router(s *gin.Engine) {
	ginAdapter := ginadapter.New(s, loggerAdapter)
	s.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	for _, m := range getControllers() {
		m.RegisterRoutes(ginAdapter)
	}

	setupErrorHandlers(ginAdapter)
}

func getControllers() []httpgateway.Routable {
	// Récupérer la base de données MongoDB
	mongoDatabase := dbAdapter.GetClient().(*mongo.Client).Database(env.Get("DB_NAME"))

	// Initialiser les repositories
	nsevenRepo := nsevenrepository.NewMongoNsevenRepository(mongoDatabase)

	// Initialiser les use cases
	nsevenUseCase := nsevenusecase.NewNsevenUseCase(nsevenRepo)

	// Retourner les controllers
	return []httpgateway.Routable{
		testcontroller.New(),
		nsevencontroller.New(nsevenUseCase),
	}
}

func setupErrorHandlers(r httpgateway.Router) {
	r.NoMethod(func(ctx httpgateway.Context) {
		ctx.MethodNotAllowed("Méthode non autorisée.", "Méthode non autorisée.")
	})
	r.NoRoute(func(ctx httpgateway.Context) {
		ctx.NotFound("Route inconnue.", "Route inconnue.")
	})
}

func getCorsConfig() cors.Config {
	return cors.Config{
		AllowOrigins:     []string{env.Get("CORS_DEV_APP"), env.Get("CORS_PREPROD_APP"), env.Get("CORS_PROD_APP")},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

func setCors(s *gin.Engine) {
	s.Use(cors.New(getCorsConfig()))
}

func setSwaggerOpt(hostTraefikApi string) {
	docs.SwaggerInfo.Host = hostTraefikApi
}

func infoServer(hostTraefikApi string) {
	loggerAdapter.If("Lancement du serveur : https://%%v", hostTraefikApi)
	loggerAdapter.If("Lancement du Swagger : https://%%v/swagger/index.html", hostTraefikApi)
}

func initDatabase(ctx context.Context) {
	if err := dbAdapter.Connect(ctx); err != nil {
		loggerAdapter.Ef("Impossible de se connecter à MongoDB : %%v", err)
		os.Exit(1)
	}
}

func closeDatabase(ctx context.Context) {
	if err := dbAdapter.Disconnect(ctx); err != nil {
		loggerAdapter.Ef("Erreur lors de la déconnexion de MongoDB : %%v", err)
	}
}

func startServerWithGracefulShutdown(s *gin.Engine, host, port string) {
	srv := &http.Server{
		Addr:    host + ":" + port,
		Handler: s,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggerAdapter.Ef("Erreur serveur : %%v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	loggerAdapter.If("Signal reçu, arrêt en cours...")
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		loggerAdapter.Ef("Erreur shutdown : %%v", err)
	}
}

func extractStringInBacktick(s string) string {
	start := strings.Index(s, "` + "`" + `")
	end := strings.LastIndex(s, "` + "`" + `")

	if start == -1 || end == -1 || start == end {
		return ""
	}

	return s[start+1 : end]
}
`, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName)
}
