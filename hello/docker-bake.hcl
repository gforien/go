target "default" {
  context    = "."
  dockerfile = "hello.Dockerfile"
  tags       = ["docker.io/gforien/hello-go:latest"]
  platforms  = ["linux/amd64", "linux/arm64"]
}
