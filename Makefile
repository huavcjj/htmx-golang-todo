PACKAGES := $(shell go list ./...)
name := $(shell basename ${PWD})

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: init
init:
	go mod init ${module}
	go install github.com/cosmtrek/air@latest
	asdf reshim golang

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: test
test:
	go test -race -cover $(PACKAGES)

.PHONY: build
build: test
	go build -o ./app -v

.PHONY: docker-build
docker-build: test
	GOPROXY=direct docker buildx build -t ${name} .

.PHONY: docker-run
docker-run:
	docker run -it --rm -p 8080:8080 ${name}

.PHONY: start
start: build
	air

.PHONY: css
css:
	tailwindcss -i css/input.css -o css/output.css --minify

.PHONY: css-watch
css-watch:
	tailwindcss -i css/input.css -o css/output.css --watch