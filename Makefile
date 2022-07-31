.PHONY: all push build deploy

all: build push deploy

push:
	@kim push dukeman/wiotemp

build:
	@kim build -t dukeman/wiotemp:latest .

buildx:
	docker buildx build \
		--push \
		--platform linux/arm/v7,linux/arm64/v8,linux/amd64 \
		--tag dukeman/wioc02:latest .

deploy:
	@echo "Deploying manifest"
	kubectl kustomize | kubectl apply -f -

kim-install:
	@echo "Installing kim"
	kim builder