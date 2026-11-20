# Paths to tools needed in dependencies
GO := $(shell which go)

# Paths to locations, etc
BUILD_DIR := "build"
EXAMPLES_DIR := $(wildcard examples/*)

# Targets
all: clean examples

examples: $(EXAMPLES_DIR)

test: go-dep
	@echo Test
	@${GO} mod tidy
	@${GO} test ./...

$(EXAMPLES_DIR): go-dep mkdir
	@echo Build example $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ./$@

FORCE:

go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean

.PHONY: examples $(EXAMPLES_DIR)
