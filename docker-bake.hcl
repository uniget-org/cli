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

target "test" {
    target = "unit-test"
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