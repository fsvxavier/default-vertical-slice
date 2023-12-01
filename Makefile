export 
$(grep -v '^#' .env | xargs)
LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)
PATH=$PATH:$(go env GOPATH)/bin




# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


swag-v1: ### swag init
	swag init -g internal/controller/http/v1/router.go
.PHONY: swag-v1

run: swag-v1 ### swag run
	go mod tidy && go mod download && \
	DISABLE_SWAGGER_HTTP_HANDLER='' GIN_MODE=debug CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### check by hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ### check by dotenv linter
	dotenv-linter
.PHONY: linter-dotenv

test: ### run test
	go test -v -cover -race ./internal/...
.PHONY: test

integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

mock: ### run mockgen
	mockgen -source ./internal/usecase/interfaces.go -package usecase_test > ./internal/usecase/mocks_test.go
.PHONY: mock

bin-deps:
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@latest


#=================================================================================================================

genmock:
	mockery --all --keeptree --dir=internal --output=./internal/mocks
.PHONY: genmock

cover:	
	go test ./... -coverprofile coverange/cover.out
	go tool cover -html=coverange/cover.out
.PHONY: cover

genproto:
	protoc --proto_path=internal/pb/proto  --go_out=internal/pb/gen --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:internal/pb/gen --go-grpc_opt=paths=source_relative internal/pb/proto/*.proto
.PHONY: genproto


runbench:
	go test -bench=. -count 5
.PHONY: runbench

#go install github.com/xo/xo@latest
xo-models:
	~/go/bin/xo schema postgres://docker:docker@localhost:5434/docker?sslmode=disable&search_path=exchange_rate -o models --srv
.PHONY: xo-models

#go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec:
	~/go/bin/gosec -exclude-dir=rules -exclude-dir=vendor -fmt=json -out=results.json -stdout ./...
.PHONY: gosec

#go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
gocritic:
	~/go/bin/gocritic check -enableAll -disable='#experimental' ./...
.PHONY: gocritic

#go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golint-html:
	~/go/bin/golangci-lint --enable=gocritic --out-format=html run ./... >> ./lint/lint.html
.PHONY: golint-html

golint-xml:
	~/go/bin/golangci-lint --enable=gocritic --out-format=checkstyle run ./... >> ./lint/lint.xml
.PHONY: golint-xml

golint-getlinters:
	~/go/bin/golangci-lint help linters
.PHONY: golint-getlinters

#go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
gofieldalt:
	~/go/bin/fieldalignment -fix ./... 
.PHONY: gofieldalt

#go install github.com/4meepo/tagalign/cmd/tagalign@latest
tagalign:
	~/go/bin/tagalign -fix ./...
.PHONY: tagalign

#go install github.com/vburenin/ifacemaker@latest
gen-interfaces:
	go generate ./...
.PHONY: gen-interfaces

#go install golang.org/x/vuln/cmd/govulncheck@latest
vuln:
	~/go/bin/govulncheck ./...
.PHONY: vuln

run-race:
	@GORACE="log_path={$PWD}/race_report.txt" go run -race main.go
.PHONY: run-race


go-gen:
	go generate ./...
.PHONY: go-gen


# stop all containers:
docker-stop-all:
	sudo docker kill $(sudo docker ps -q)
.PHONY: docker-stop-all

# Remove all containers and docker images
docker-clean-all:
	sudo docker rm $(sudo docker ps -a -q) --force
	sudo docker rmi $(sudo docker images -q) --force
	sudo docker rmi $(sudo docker images -q) --force
	sudo docker network rm $(sudo docker network ls -q)
.PHONY: docker-clean-all

docker-stats:
	sudo docker stats
.PHONY: docker-stats

docker-start:
	sudo docker-compose down
	sudo docker-compose up -d --build
.PHONY: docker-start
