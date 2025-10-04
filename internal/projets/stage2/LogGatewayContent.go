package stage2

func LogGatewayContent() string {
	return `package loggateway

type Logger interface {
	// With creates a child logger with additional key-value pairs.
	With(kv ...any) Logger

	// S = Success, I = Info, W = Warning, E = Error, F = Fatal
	S(msg string, args ...any)
	I(msg string, args ...any)
	W(msg string, args ...any)
	E(msg string, args ...any)
	F(msg string, args ...any)

	// Sf = Successf, If = Infof, Wf = Warningf, Ef = Errorf, Ff = Fatalf
	// sprintf there args
	Sf(format string, args ...any)
	If(format string, args ...any)
	Wf(format string, args ...any)
	Ef(format string, args ...any)
	Ff(format string, args ...any)

	// Close the logger and free resources
	Close()
}
`
}
