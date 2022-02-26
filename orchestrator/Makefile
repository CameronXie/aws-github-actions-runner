PWD=`pwd`
DIST=_dist
TESTS=${DIST}/tests
PKG=${PWD}/pkg/...
INTERNAL=${PWD}/internal/...
EC2_LAUNCHERS=ubuntu

build: install-dependency
	@rm -rf ${DIST}
	@make build-launcher
	@make build-terminator

compile-launcher:
	@go build -o ${DIST}/launcher/ec2/launcher cmd/launcher/ec2/main.go
	@go build -o ${DIST}/launcher/eks/launcher cmd/launcher/eks/main.go

build-launcher: compile-launcher
	@for l in $(EC2_LAUNCHERS); do cp -r cmd/launcher/ec2/$$l ${DIST}/launcher/ec2/$$l && \
		cp ${DIST}/launcher/ec2/launcher ${DIST}/launcher/ec2/$$l/launcher; done
	@rm ${DIST}/launcher/ec2/launcher

build-terminator:
	@go build -o ${DIST}/terminator/ec2/terminator cmd/terminator/ec2/main.go
	@go build -o ${DIST}/terminator/eks/terminator cmd/terminator/eks/main.go

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