package stage2

import "fmt"

func ReleasercRootContent() string {
	return fmt.Sprintf(`{
  "branches": ["prod"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    ["@semantic-release/release-notes-generator", { "preset": "conventionalcommits" }],
    ["@semantic-release/github", { "assets": [] }]
  ]
}
`)
}