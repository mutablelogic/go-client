# Paths to tools needed in dependencies
GO := $(shell which go)
NPM := $(shell which npm)

# Build flags
BUILD_MODULE := $(shell go list -m)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

# Paths to locations, etc
BUILD_DIR := "build"
PLUGIN_DIR := $(filter-out $(wildcard plugin/*.go), $(wildcard plugin/*))
NPM_DIR := $(wildcard npm/*)
CMD_DIR := $(wildcard cmd/*)

# Targets
all: clean cmd npm plugins

cmd: $(CMD_DIR)

plugins: $(PLUGIN_DIR)

npm: $(NPM_DIR)

test:
	@${GO} mod tidy
	@${GO} test -v ./pkg/...

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

$(PLUGIN_DIR): dependencies mkdir
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

$(NPM_DIR): dependencies-npm
	@echo Build npm $(notdir $@)
	@cd $@ && npm install && npm run build

FORCE:

dependencies:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

dependencies-npm:
	@test -f "${NPM}" && test -x "${NPM}" || (echo "Missing npm binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@find ${NPM_DIR} -name dist -type d -prune -exec rm -fr {} \;
	@${GO} mod tidy
	@${GO} clean

distclean: clean
	@find ${NPM_DIR} -name node_modules -type d -prune -exec rm -fr {} \;
