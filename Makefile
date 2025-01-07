.PHONY: tidy build test

# directories
ROOT_DIR = $(CURDIR)
OUTPUT_DIR = $(ROOT_DIR)/build/$(VERSION)

# version
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITTAG    = $(shell git describe --tags --always)
GITHASH   = $(shell git rev-parse --short HEAD)
VERSION  ?= ${GITTAG}-$(shell date +%y.%m.%d)

# ldflags
# output directory for release package and version for command line
LDVersionFLAG = "-X git.woa.com/bk-gse/bk-nodeman/internal/version.VERSION=${VERSION} \
    	-X git.woa.com/bk-gse/bk-nodeman/internal/version.BUILDTIME=${BUILDTIME} \
    	-X git.woa.com/bk-gse/bk-nodeman/internal/version.GITHASH=${GITHASH}"


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
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-backend $(ROOT_DIR)/cmd/backend/main.go

saas: pre
	CGO_ENABLED=0 go build -ldflags ${LDVersionFLAG} -o $(OUTPUT_DIR)/bk-nodeman-saas $(ROOT_DIR)/cmd/saas/*.go

docker-build: backend saas
	$(CP) $(ROOT_DIR)/install/images/Dockerfile $(OUTPUT_DIR)
	$(CD) $(OUTPUT_DIR) && docker build -t bk-nodeman:v${VERSION} .

all: backend saas

clean:
	$(RM) -rf build