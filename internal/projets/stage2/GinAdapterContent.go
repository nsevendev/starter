package stage2

import "fmt"

func GinAdapterContent(moduleName string) string {
	return fmt.Sprintf(`package ginadapter

import (
	"github.com/gin-gonic/gin"
	"github.com/nsevenpack/ginresponse"
	"mime/multipart"
	"net/http"
	"%s/internal/application/gateway/httpgateway"
	"%s/internal/application/gateway/loggateway"
)

const defaultPrefix = "/api/v1"

type ginRouter struct {
	r      *gin.Engine
	prefix string
	logger loggateway.Logger
}

type ginContext struct {
	c      *gin.Context
	logger loggateway.Logger
}

func New(r *gin.Engine, logger loggateway.Logger) httpgateway.Router {
	ginresponse.SetFormatter(&ginresponse.JsonFormatter{})
	return &ginRouter{r: r, prefix: defaultPrefix, logger: logger}
}

// wrapHandler wraps gin.Context with logging metadata
func (g *ginRouter) wrapHandler(c *gin.Context, h func(httpgateway.Context)) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	reqLog := g.logger.With(
		"method", c.Request.Method,
		"path", path,
		"ip", c.ClientIP(),
	)
	h(&ginContext{c: c, logger: reqLog})
}

func (g *ginRouter) Handle(method, path string, h func(httpgateway.Context)) {
	path = g.prefix + path
	g.r.Handle(method, path, func(c *gin.Context) {
		g.wrapHandler(c, h)
	})
}

func (g *ginRouter) NoRoute(h func(httpgateway.Context)) {
	g.r.NoRoute(func(c *gin.Context) {
		g.wrapHandler(c, h)
	})
}

func (g *ginRouter) NoMethod(h func(httpgateway.Context)) {
	g.r.NoMethod(func(c *gin.Context) {
		g.wrapHandler(c, h)
	})
}

/* CONTEXT METHOD */

func (g *ginContext) Abort() {
	g.c.Abort()
}

func (g *ginContext) Request() *http.Request {
	return g.c.Request
}

func (g *ginContext) Param(name string) string {
	return g.c.Param(name)
}

func (g *ginContext) Query(name string) string {
	return g.c.Query(name)
}

func (g *ginContext) Header(name string) string {
	return g.c.GetHeader(name)
}

func (g *ginContext) SetHeader(name, value string) {
	g.c.Header(name, value)
}

func (g *ginContext) Logger() loggateway.Logger {
	return g.logger
}

func (g *ginContext) JSON(code int, obj any) {
	g.c.JSON(code, obj)
}

/* METHODE FOR FILE */

// FormFile File upload methods
func (g *ginContext) FormFile(name string) (*multipart.FileHeader, error) {
	return g.c.FormFile(name)
}

func (g *ginContext) SaveUploadedFile(f *multipart.FileHeader, dst string) error {
	return g.c.SaveUploadedFile(f, dst)
}

/* SUCCESS REPONSE */

// Success responses (2xx)
func (g *ginContext) Success(message string, data any) {
	g.logAndRespond(g.logger.Sf, message, data, ginresponse.Success)
}

func (g *ginContext) Created(message string, data any) {
	g.logAndRespond(g.logger.Sf, message, data, ginresponse.Created)
}

func (g *ginContext) NoContent(message string) {
	g.logger.Sf(message, nil)
	ginresponse.NoContent(g.c, message)
}

/* ERROR RESPONSE */

// BadRequest Client error responses (4xx)
func (g *ginContext) BadRequest(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.BadRequest)
}

func (g *ginContext) Unauthorized(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.Unauthorized)
}

func (g *ginContext) Forbidden(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.Forbidden)
}

func (g *ginContext) NotFound(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.NotFound)
}

func (g *ginContext) UnprocessableEntity(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.UnprocessableEntity)
}

func (g *ginContext) Conflict(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.Conflict)
}

func (g *ginContext) TooManyRequests(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.TooManyRequests)
}

func (g *ginContext) MethodNotAllowed(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.MethodNotAllowed)
}

func (g *ginContext) NotAcceptable(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.NotAcceptable)
}

// ServiceUnavailable Server error responses (5xx)
func (g *ginContext) ServiceUnavailable(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.ServiceUnavailable)
}

func (g *ginContext) InternalServerError(message string, err any) {
	g.logAndRespond(g.logger.Ef, message, err, ginresponse.InternalServerError)
}

/* CENTRALISATION DES LOG ET RESPONSE */

// logAndRespond centralizes logging and response logic
func (g *ginContext) logAndRespond(
	logFn func(string, ...any),
	message string,
	data any,
	responseFn func(*gin.Context, string, any),
) {
	logFn(message, data)
	responseFn(g.c, message, data)
}
`, moduleName, moduleName)
}
