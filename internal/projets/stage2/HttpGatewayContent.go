package stage2

import "fmt"

func HttpGatewayContent(moduleName string) string {
	return fmt.Sprintf(`package httpgateway

import (
	"mime/multipart"
	"net/http"
	"%s/internal/application/gateway/loggateway"
)

type ErrorResponse struct {
	Message string ` + "`json:\"message\"`" + `
	Type    string ` + "`json:\"type\"`" + `
	Field   string ` + "`json:\"field\"`" + `
	Detail  string ` + "`json:\"detail\"`" + `
}

type Context interface {
	Request() *http.Request
	Param(name string) string
	Query(name string) string
	Header(name string) string
	SetHeader(name, value string)
	Abort()

	Success(message string, data any)
	Created(message string, data any)
	NoContent(message string)
	BadRequest(message string, err any)
	Unauthorized(message string, err any)
	Forbidden(message string, err any)
	NotFound(message string, err any)
	UnprocessableEntity(message string, err any)
	Conflict(message string, err any)
	TooManyRequests(message string, err any)
	MethodNotAllowed(message string, err any)
	NotAcceptable(message string, err any)
	ServiceUnavailable(message string, err any)
	InternalServerError(message string, err any)

	FormFile(name string) (*multipart.FileHeader, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error

	Logger() loggateway.Logger
}

type Router interface {
	Handle(method, path string, h func(Context))
	NoRoute(h func(Context))
	NoMethod(h func(Context))
}

type Routable interface {
	RegisterRoutes(Router)
}
`, moduleName)
}
