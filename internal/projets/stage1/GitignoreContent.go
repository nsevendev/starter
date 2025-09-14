package stage1

func GitignoreRootContent() string {
	return `# Angular specific
/dist/
/out-tsc/
/tmp/
/coverage/
/e2e/test-output/
/.angular/
.angular/

.vscode
.idea
/app/.idea
/app/.vscode
/api/.idea
/api/.vscode
.DS_Store
app/.DS_Store
api/.DS_Store

# Node modules and dependency files
/node_modules/
/package-lock.json
/yarn.lock

# Environment files
/.env
app/.env
api/.env

# Angular CLI and build artefacts
/.angular-cli.json
/.ng/

# TypeScript cache
*.tsbuildinfo

# Logs
npm-debug.log*
yarn-debug.log*
yarn-error.log*
`
}
