COVERFILE = coverage.out
FILTERED_COVERFILE = coverage.filtered.out

SRC_DIRS = ./...
TEST_FLAGS = -covermode=atomic -coverprofile=$(COVERFILE)


.PHONY: mocks test filter-cover cover test-cover
mocks:
	@echo "======== Создание моков... ========"
	@mockgen -source=store_service/internal/usecase/item_usecase.go  -destination=store_service/internal/usecase/mock/mock_item_repository.go  -package=mock ItemRepository
	@mockgen -source=store_service/internal/usecase/store_usecase.go -destination=store_service/internal/usecase/mock/mock_store_repository.go -package=mock StoreRepository
	@mockgen -source=store_service/internal/delivery/http/item_handler.go  -destination=store_service/internal/delivery/mock/mock_item_usecase.go  -package=mock ItemUsecaseInterface
	@mockgen -source=store_service/internal/delivery/http/store_handler.go -destination=store_service/internal/delivery/mock/mock_store_usecase.go -package=mock StoreUsecaseInterface
	@echo "======== Моки созданы ========"

test:
	go test $(SRC_DIRS) $(TEST_FLAGS)
	@echo "======== Тесты завершены ========"

filter-cover:
	@grep -vE "mock_|_test.go" $(COVERFILE) > $(FILTERED_COVERFILE)
	@echo "======== Покрытие очищено от моков и тестов ========"

cover: test filter-cover
	@echo ""
	@echo "======== Покрытие по функциям: ========"
	@go tool cover -func=$(FILTERED_COVERFILE)
	@echo ""
	@echo "======== Общее покрытие ========"
	@go tool cover -func=$(FILTERED_COVERFILE) | grep total:


test-cover: mocks cover

# генерация сваггеров
.PHONY: swagger

swagger:
	@swag init -g cmd/store_service/main.go --output ./docs/store --ot yaml


.PHONY: build build-test

build: swagger
	@docker-compose down -v
	@docker-compose up --build

build-test: build test-cover
