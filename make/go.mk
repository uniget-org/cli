.PHONY:
run--%:
	@\
	go run ./cmd/$*

.PHONY:
run: run--docker-setup

.PHONY:
run-in-docker: run-in-docker--docker-setup

.PHONY:
run-in-docker--%:
	@\
	docker run --interactive --tty --rm \
	    --mount type=bind,src=$${HOME}/go/pkg/mod,dst=/go/pkg/mod \
	    --mount type=bind,src=$${HOME}/.cache/go-build,dst=/.cache/go-build \
		--mount type=bind,src=$${PWD},dst=/src \
		--workdir /src \
		golang:alpine \
			go run ./cmd/$*