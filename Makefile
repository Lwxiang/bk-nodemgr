MKDIR=mkdir -p

default: all
pre:
	$(MKDIR) build
	go mod tidy

backend: pre
	go build -o build/backend ./cmd/backend/backend.go

all: backend
