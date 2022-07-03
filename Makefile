COMPONENTS=producer messenger publisher orchestrator
RUNNER_ECR="${CDK_DEFAULT_ACCOUNT}.dkr.ecr.${CDK_DEFAULT_REGION}.amazonaws.com/actions-runner-ecr"
RUNNER_TAG=$(RUNNER_ECR):latest

# Docker
up: create-dev-env
	@docker compose up --build -d

down:
	@docker compose down -v

create-dev-env:
	@test -e .env || cp .env.example .env

ci-deploy:
	@make build
	@make deploy

build:
	@for c in $(COMPONENTS); do $(MAKE) build -C $$c; done
	@npm ci
	@npm run build

test:
	@for c in $(COMPONENTS); do $(MAKE) test -C $$c; done

deploy:
	@npm run cdk -- context --clear
	@npm run cdk -- bootstrap aws://${CDK_DEFAULT_ACCOUNT}/us-east-1
	@make -j 2 deploy-images deploy-eks
	@npm run cdk -- deploy actions-runner-orchestrator
	@./bin/create_runner_namespace.sh

build-runner-image:
	# build ubuntu runner
	@docker build --platform linux/amd64 --build-arg RUNNER_VERSION -t $(RUNNER_TAG) ./bin/runner/docker/ubuntu

login-ecr:
	@aws ecr get-login-password --region ${CDK_DEFAULT_REGION} | docker login --username AWS --password-stdin ${RUNNER_ECR}

push-runner-image: login-ecr build-runner-image
	@docker push $(RUNNER_TAG)


deploy-images:
	@npm run cdk -- deploy actions-runner-ecr
	@make push-runner-image

deploy-eks:
	@npm run cdk -- deploy actions-runner-vpc
	@./bin/create_eks.sh
