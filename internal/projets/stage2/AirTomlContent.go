package stage2

func AirTomlContent() string {
	return `root = "."
tmp_dir = "tmp/air"
env = ["SERVICE=api"]

[build]
cmd = "sh -c 'swag init -o docs -g cmd/${SERVICE}/main.go --parseInternal --pd && go clean -cache -modcache -testcache && go build -o ./tmp/air/${SERVICE}/main ./cmd/${SERVICE}/main.go'"
#cmd = "swag init -o docs -g main.go --parseInternal --pd && go clean && go build -o ./tmp/air/main ./main.go"
bin = "tmp/air/${SERVICE}/main"
include_ext = ["go"]
include_dir = ["cmd", "internal", "pkg"]
exclude_dir = ["tmp", "doc", "docs"]
watch_dir = "."

[log]
log = "build.log"
time = true
`
}
