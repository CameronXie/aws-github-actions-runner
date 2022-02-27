PWD=`pwd`
DIST=_dist
TESTS=${DIST}/tests
INTERNAL=${PWD}/internal/...

build: install-dependency
	@rm -rf ${DIST}
	@make build-messenger

build-messenger:
	@go build -o ${DIST}/messenger cmd/main.go

install-dependency:
	@go mod vendor

# app
test: install-dependency
	@make app-lint
	@make app-unit

app-lint:
	@golangci-lint run ${INTERNAL} -v

app-unit:
	@mkdir -p ${TESTS}
	@go clean -testcache
	@go test \
        -cover \
        -coverprofile=cp.out \
        -outputdir=${TESTS} \
        -race \
        -v \
        -failfast \
        ${INTERNAL}
	@go tool cover -html=${TESTS}/cp.out -o ${TESTS}/cp.html