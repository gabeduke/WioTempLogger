.PHONY: all push build deploy

all: build push deploy

push:
	kim push dukeman/wio-temp-hu-logger

build:
	kim build -t dukeman/wio-temp-hu-logger:latest .

deploy:
	@echo "Deploying manifest"
	kubectl kustomize | kubectl apply -f -