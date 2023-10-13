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