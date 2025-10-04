package stage2

func GitignoreApiContent() string {
	return `tmp/*
!tmp/.gitkeep
.idea
.vscode
`
}
