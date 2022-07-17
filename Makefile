.PHONY: all push build deploy

all: build push deploy

push:
	@kim push dukeman/wiotemp

build:
	@kim build -t dukeman/wiotemp:latest .

deploy:
	@echo "Deploying manifest"
	kubectl kustomize | kubectl apply -f -