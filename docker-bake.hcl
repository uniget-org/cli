target "binary" {
    target = "bin-linux"
    args = {
        version = "main"
    }
    output = [
        "type=local,dest=."
    ]
}