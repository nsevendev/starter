package stage2

func PackageJsonScriptContent() map[string]string {
	return map[string]string{
		"dev":     "astro dev --host 0.0.0.0 --port 3000 --poll 2000",
		"build":   "astro check && astro build",
		"preview": "astro preview",
		"astro":   "astro",
		"check":   "astro check",
	}
}
