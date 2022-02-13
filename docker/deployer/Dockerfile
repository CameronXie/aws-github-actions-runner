FROM cameronx/aws-deployer:latest

ARG GOLANG_VERSION=1.17.6
ARG GOLANGCI_LINT_VERSION=1.44.0

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$GOPATH/bin:$PATH
RUN set -eux; \
	arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; url=; \
	case "$arch" in \
		'amd64') url="https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz" ;; \
		'arm64') url="https://dl.google.com/go/go${GOLANG_VERSION}.linux-arm64.tar.gz" ;; \
		*) echo >&2 "error: unsupported architecture '$arch'"; exit 1 ;; \
	esac \
	&& build_dir="$(mktemp -d)" && cd "${build_dir}" \
	&& curl -o go.tgz.asc -L "$url.asc" \
	&& curl -o go.tgz -L "$url" \
    && GNUPGHOME="$(mktemp -d)"; export GNUPGHOME \
	&& gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796' \
	&& gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13' \
	&& gpg --batch --verify go.tgz.asc go.tgz \
	&& gpgconf --kill all \
    && rm -rf "$GNUPGHOME" go.tgz.asc \
    && tar -C /usr/local -xzf go.tgz \
	&& rm -rf "${build_dir}" \
    && go version

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
      sh -s -- -b $(go env GOPATH)/bin v${GOLANGCI_LINT_VERSION} \
    && npm install -g aws-cdk
