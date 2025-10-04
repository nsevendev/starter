package stage2

import "fmt"

func LoggerAdapterContent(moduleName string) string {
	return fmt.Sprintf(`package loggeradapter

import (
	"fmt"
	base "github.com/nsevenpack/logger/v2/logger"
	"%s/internal/application/gateway/loggateway"
	"slices"
	"strings"
)

type loggerAdapter struct {
	fields map[string]any // champs de contexte (request_id, user_id, path, ...)
}

// New initialise ton logger de base et retourne l'adapter.
func New(appEnv string) loggateway.Logger {
	base.Init(appEnv)
	return &loggerAdapter{fields: map[string]any{}}
}

// With ajoute/écrase des champs et retourne un nouveau logger enrichi.
func (l *loggerAdapter) With(kv ...any) loggateway.Logger {
	m := cloneMap(l.fields)
	mergeKV(m, kv...)
	return &loggerAdapter{fields: m}
}

// ---------- Public API : non formaté ----------

func (l *loggerAdapter) S(msg string, args ...any) {
	base.S(prefix(l.fields) + sprintArgs(msg, args...))
}
func (l *loggerAdapter) I(msg string, args ...any) {
	base.I(prefix(l.fields) + sprintArgs(msg, args...))
}
func (l *loggerAdapter) W(msg string, args ...any) {
	base.W(prefix(l.fields) + sprintArgs(msg, args...))
}
func (l *loggerAdapter) E(msg string, args ...any) {
	base.E(prefix(l.fields) + sprintArgs(msg, args...))
}
func (l *loggerAdapter) F(msg string, args ...any) {
	base.F(prefix(l.fields) + sprintArgs(msg, args...))
}

// ---------- Public API : formaté ----------

func (l *loggerAdapter) Sf(format string, args ...any) { base.Sf(prefix(l.fields)+format, args...) }
func (l *loggerAdapter) If(format string, args ...any) { base.If(prefix(l.fields)+format, args...) }
func (l *loggerAdapter) Wf(format string, args ...any) { base.Wf(prefix(l.fields)+format, args...) }
func (l *loggerAdapter) Ef(format string, args ...any) { base.Ef(prefix(l.fields)+format, args...) }
func (l *loggerAdapter) Ff(format string, args ...any) { base.Ff(prefix(l.fields)+format, args...) }

// ---------- Lifecycle ----------

func (l *loggerAdapter) Close() { base.Close() }

// ---------- Helpers internes ----------

func prefix(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	// on trie pour des logs déterministes
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var b strings.Builder
	b.WriteString("[")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(fmt.Sprintf("%%v", m[k]))
	}
	b.WriteString("] ")
	return b.String()
}

// pour S/I/W/E/F (non formaté) : on concatène msg + args optionnels
func sprintArgs(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, msg)
	for _, a := range args {
		parts = append(parts, fmt.Sprintf("%%v", a))
	}
	return strings.Join(parts, " ")
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeKV(dst map[string]any, kv ...any) {
	for i := 0; i+1 < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			continue
		}
		dst[k] = kv[i+1]
	}
	// si impair, on garde une trace (debug)
	if len(kv)%%2 == 1 {
		dst["_kv_last_unpaired"] = fmt.Sprintf("%%v", kv[len(kv)-1])
	}
}
`, moduleName)
}
