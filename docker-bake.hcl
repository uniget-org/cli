variable "distro" {
    default = "alpine-3.15"
}

target "default" {
    context = "."
    dockerfile = "test/Dockerfile.${distro}"
    tags = [
        "nicholasdille/docker-setup:${distro}"
    ]
}