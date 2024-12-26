# directories
ROOT_DIR = $(CURDIR)
OUTPUT_DIR = $(ROOT_DIR)/build/$(VERSION)

# version
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITTAG    = $(shell git describe --tags --always)
GITHASH   = $(shell git rev-parse --short HEAD)
VERSION  ?= ${GITTAG}-$(shell date +%y.%m.%d)

# cmd
MKDIR = mkdir -p
ECHO  = $(if $(filter Linux,$(shell uname)),echo -e,echo)
CD    = cd
CP    = cp
RM    = rm
SH    = sh

default: all
pre:
	$(MKDIR) $(OUTPUT_DIR)
	go mod tidy

backend: pre
	go build -o $(OUTPUT_DIR)/bk-nodeman-backend $(ROOT_DIR)/cmd/backend/backend.go

images: backend
	$(CP) $(ROOT_DIR)/install/images/Dockerfile $(OUTPUT_DIR)/
	$(CD) $(OUTPUT_DIR) && docker build -t bk-nodeman:v${VERSION} .

all: backend

clean:
	$(RM) -rf build