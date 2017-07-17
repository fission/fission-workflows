# go_rules

git_repository(
	name = "io_bazel_rules_go",
	remote = "https://github.com/bazelbuild/rules_go.git",
	tag = "0.5.1",
)
load("@io_bazel_rules_go//go:def.bzl", "go_repositories")

go_repositories(
	go_version = "1.8.3",
)

# Docker
git_repository(
	name = "io_bazel_rules_docker",
	remote = "https://github.com/bazelbuild/rules_docker.git",
	tag = "v0.0.2",
)
load("@io_bazel_rules_docker//docker:docker.bzl", "docker_repositories")

docker_repositories()
