target "binary" {
    target = "bin-linux"
    args = {
        version = "main"
    }
    output = [
        "type=local,dest=."
    ]
}

target "lint" {
    target = "lint"
}

target "vet" {
    target = "vet"
}

target "test" {
    target = "unit-test"
}

target "gosec" {
    target = "gosec"
}

target "staticcheck" {
    target = "staticcheck"
}

target "cli-test" {
    target = "cli-test"
}

group "full" {
    targets = [
        "binary",
        "lint",
        "vet",
        "test",
        "gosec",
        "staticcheck",
        "cli-test"
    ]
}