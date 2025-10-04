package stage2

import (
	"fmt"
	"strings"
)

func AstroConfigContent(portLinkTraefik int, allowedHost []string) string {
	hostsFormatted := formatJSArray(allowedHost)

	return fmt.Sprintf(`// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from "@tailwindcss/vite";
import node from '@astrojs/node';

// https://astro.build/config
export default defineConfig({
  output: 'server',
  adapter: node({
    mode: 'standalone'
  }),
  vite: {
    plugins: [tailwindcss()],
    server: {
      allowedHosts: %s
    }
  },
  server: {
    host: '0.0.0.0',
    port: %v
  }
});
`, hostsFormatted, portLinkTraefik)
}

func formatJSArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("'%s'", item)
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}
