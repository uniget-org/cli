variable "distro" {
    default = "alpine-3.15"
}

target "default" {
    context = "."
    dockerfile = "env/${distro}/Dockerfile"
    tags = [
        "nicholasdille/docker-setup:${distro}"
    ]
}