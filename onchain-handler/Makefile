.PHONY: onchain-handler

build: onchain-handler
onchain-handler:
	go build -o ./onchain-handler ./cmd/main.go
clean:
	rm -i -f onchain-handler

migrate:
	go run cmd/migration/main.go	

run-test:
	go test -v ./internal/infra/caching/test
	go test -v ./internal/util/test
	go test -v ./test

restart: stop clean build start
	@echo "onchain-handler restarted!"

build-service: clean build
	@echo "Restart service with cmd: 'systemctl restart onchain-handler'"
	systemctl restart onchain-handler

run: build
	@echo "Starting the onchain-handler..."
	@env DB_PASSWORD=${DB_PASSWORD} ./onchain-handler &
	@echo "onchain-handler running!"

stop:
	@echo "Stopping the onchain-handler..."
	@-pkill -SIGTERM -f "onchain-handler"
	@echo "Stopped onchain-handler"

lint:
	golangci-lint run --fix

swagger:
	swag init -g cmd/main.go