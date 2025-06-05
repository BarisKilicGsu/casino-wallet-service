.PHONY: build run test mock clean

# Go parametreleri
BINARY_NAME=casino-wallet-service
MAIN_PATH=./cmd/api

# Docker parametreleri
DOCKER_COMPOSE=docker compose

# Mockery kurulumu
MOCKERY_VERSION=v2.42.1


tools/swagger:
	$(call print-target)
	GOBIN=$(CURDIR)/tools go install github.com/go-swagger/go-swagger/cmd/swagger@v0.27.0

.PHONY: install_tools
install_tools: tools/swagger

.PHONY: models
models: tools/swagger
	$(call print-target)
	find ./models -type f -delete
	./tools/swagger  generate model DEBUG=1 --spec=docs/api.yaml

.PHONY: mock
mock: $(call print-target)
	rm -rf ./mocks/*
	mockery --all

define print-target
	@printf "Executing target: \033[36m$@\033[0m\n"
endef



