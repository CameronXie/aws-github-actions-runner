PWD=`pwd`
DIST=_dist
TESTS=${DIST}/tests
PKG=${PWD}/pkg/...
INTERNAL=${PWD}/internal/...

build: install-dependency
	@rm -rf ${DIST}
	@make build-launcher
	@make build-terminator

build-launcher:
	@go build -o ${DIST}/launcher/launcher cmd/launcher/main.go
	@cp -r cmd/launcher/userdata ${DIST}/launcher/

build-terminator:
	@go build -o ${DIST}/terminator/terminator cmd/terminator/main.go

install-dependency:
	@go mod vendor

# app
test: install-dependency
	@make app-lint
	@make app-unit

app-lint:
	@golangci-lint run ${PKG} ${INTERNAL} -v

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
        ${PKG} ${INTERNAL}
	@go tool cover -html=${TESTS}/cp.out -o ${TESTS}/cp.html