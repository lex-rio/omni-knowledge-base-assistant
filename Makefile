BINARY := omni
CMD := ./cmd/omni
BUILD_FLAGS := -trimpath -ldflags="-s -w"

.PHONY: build run test lint clean cross-arm64

build:
	go build $(BUILD_FLAGS) -o $(BINARY) $(CMD)

run:
	go run $(CMD)

test:
	go test ./... -count=1

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

cross-arm64:
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY)-arm64 $(CMD)
