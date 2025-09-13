package stage1

// ReleasercContent retourne le contenu standard d'un fichier .releaserc.json
func ReleasercContent() string {
	return `
{
  "branches": ["main"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    ["@semantic-release/release-notes-generator", { "preset": "conventionalcommits" }],
    ["@semantic-release/github", { "assets": [] }]
  ]
}
`
}
