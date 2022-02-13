FROM ubuntu:20.04

ARG USER_NAME=runner
ARG USER_ID=1001

ARG RUNNER_VERSION
ARG GOLANG_VER=1.17.6
ARG DOCKER_COMPOSE_VERSION=1.29.2
ARG NODE_VERSION=16

ENV RUNNER_HOMEDIR=/home/${USER_NAME}
ENV RUNNER_WORKDIR=${RUNNER_HOMEDIR}/work
ENV DEBIAN_FRONTEND=noninteractive

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

USER root
WORKDIR /tmp/build

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
       apt-transport-https \
       ca-certificates \
       curl \
       dumb-init \
       git \
       gnupg \
       lsb-release \
       jq \
       sudo \
       tar \
       unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /
RUN groupadd --gid ${USER_ID} ${USER_NAME} \
    && useradd --uid ${USER_ID} --gid ${USER_NAME} --shell /bin/bash --create-home ${USER_NAME} \
    && echo "${USER_NAME} ALL=(ALL) NOPASSWD: /usr/bin/apt-get" >> /etc/sudoers \
    && mkdir -p ${RUNNER_WORKDIR} \
    && chown -R ${USER_NAME}:${USER_NAME} /entrypoint.sh ${RUNNER_WORKDIR} \
    && chmod +x /entrypoint.sh

# install Docker
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update && apt-get install -y docker-ce docker-ce-cli containerd.io \
    && usermod -aG docker ${USER_NAME} \
    && curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose \
    && chown root:docker /usr/local/bin/docker-compose

# install nodejs
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs

# install Golang
ENV PATH /usr/local/go/bin:$PATH
RUN set -eux; \
	arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; url=; \
	case "$arch" in \
		'amd64') url="https://dl.google.com/go/go${GOLANG_VER}.linux-amd64.tar.gz" ;; \
		'arm64') url="https://dl.google.com/go/go${GOLANG_VER}.linux-arm64.tar.gz" ;; \
		*) echo >&2 "error: unsupported architecture '$arch'"; exit 1 ;; \
	esac \
	&& curl -o go.tgz.asc -L "$url.asc" \
	&& curl -o go.tgz -L "$url" \
	&& gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796' \
    && gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13' \
    && gpg --batch --verify go.tgz.asc go.tgz \
	&& gpgconf --kill all \
    && tar -C /usr/local -xzf go.tgz

# clean up
RUN rm -rf /tmp/build

USER ${USER_ID}
WORKDIR ${RUNNER_HOMEDIR}

# install runner
RUN arch="$(dpkg --print-architecture)" && arch="${arch##*-}" \
    && case "$arch" in \
      'amd64') runner_arch="x64" ;; \
      'arm64') runner_arch="arm64" ;; \
      *) echo >&2 "error: unsupported architecture '$arch'"; exit 1 ;; \
    esac \
    && curl -o ${RUNNER_HOMEDIR}/actions-runner-linux.tar.gz -L https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-${runner_arch}-${RUNNER_VERSION}.tar.gz \
    && tar xzf ${RUNNER_HOMEDIR}/actions-runner-linux.tar.gz \
    && rm ${RUNNER_HOMEDIR}/actions-runner-linux.tar.gz

ENTRYPOINT ["/entrypoint.sh"]
CMD ["./bin/Runner.Listener", "run", "--startuptype", "service"]
