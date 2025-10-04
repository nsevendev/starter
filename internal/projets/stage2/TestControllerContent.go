package stage2

import "fmt"

func TestControllerContent(moduleName string) string {
	return fmt.Sprintf(`package testcontroller

import (
	"%s/internal/application/gateway/httpgateway"
)

type controller struct {
	prefixUrl string
}

type Interface interface {
	SayHello(c httpgateway.Context)
	SayHelloWithDto(c httpgateway.Context)
	RegisterRoutes(r httpgateway.Router)
}

func New() Interface {
	return &controller{
		prefixUrl: "/test",
	}
}

func (ctr *controller) RegisterRoutes(r httpgateway.Router) {
	r.Handle("GET", ctr.prefixUrl+"/sayhello", ctr.SayHello)
	r.Handle("GET", ctr.prefixUrl+"/sayhellodto", ctr.SayHelloWithDto)
}
`, moduleName)
}

func TestSayHelloContent(moduleName string) string {
	return fmt.Sprintf(`package testcontroller

import (
	"%s/internal/application/gateway/httpgateway"
)

func (ctr *controller) SayHello(ctx httpgateway.Context) {
	ctx.Success("Hello test", map[string]any{"message": "Je dis bonjour"})
}

func (ctr *controller) SayHelloWithDto(ctx httpgateway.Context) {
	type helloDto struct {
		Name    string `+"`json:\"name\"`"+`
		Age     int    `+"`json:\"age\"`"+`
		Message string `+"`json:\"message,omitempty\"`"+`
	}

	ctx.Success("Hello test", helloDto{Name: "Jeanne", Age: 30, Message: "Je dis bonjour avec un DTO"})
}
`, moduleName)
}
