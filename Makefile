
EXECUTABLE := qitmeer
GITVERSION := $(shell git rev-parse --short HEAD)
DEV=dev
RELEASE=release
LDFLAG_DEV = -X github.com/Qitmeer/qitmeer/version.Build=$(DEV)-$(GITVERSION)
LDFLAG_RELEASE = -X github.com/Qitmeer/qitmeer/version.Build=$(RELEASE)
GOFLAGS_DEV = -ldflags "$(LDFLAG_DEV)"
GOFLAGS_RELEASE = -ldflags "$(LDFLAG_RELEASE)"
VERSION=$(shell ./build/bin/qitmeer --version | grep ^qitmeer | cut -d' ' -f3|cut -d'+' -f1)
GOBIN = ./build/bin

UNIX_EXECUTABLES := \
	build/release/darwin/amd64/bin/$(EXECUTABLE) \
	build/release/linux/amd64/bin/$(EXECUTABLE)
WIN_EXECUTABLES := \
	build/release/windows/amd64/bin/$(EXECUTABLE).exe

EXECUTABLES=$(UNIX_EXECUTABLES) $(WIN_EXECUTABLES)
	
COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.tar.gz) $(WIN_EXECUTABLES:%.exe=%.zip)

RELEASE_TARGETS=$(EXECUTABLES) $(COMPRESSED_EXECUTABLES)

.PHONY: qitmeer qx release

qitmeer: qitmeer-build
	@echo "Done building."
	@echo "  $(shell $(GOBIN)/qitmeer --version))"
	@echo "Run \"$(GOBIN)/qitmeer\" to launch."

qitmeer-build:
	@go build -o $(GOBIN)/qitmeer $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
qx:
	@go build -o $(GOBIN)/qx "github.com/Qitmeer/qitmeer/cmd/qx"

checkversion: qitmeer-build
#	@echo version $(VERSION)

all: qitmeer-build qx

# amd64 release
build/release/%: OS=$(word 3,$(subst /, ,$(@)))
build/release/%: ARCH=$(word 4,$(subst /, ,$(@)))
build/release/%/$(EXECUTABLE):
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
build/release/%/$(EXECUTABLE).exe:
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"

%.zip: %.exe
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).zip "$<"
%.tar.gz : %
	@echo tar $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@tar -zcvf $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).tar.gz "$<"
release: clean checkversion
	@echo "Build release version : $(VERSION)"
	@$(MAKE) $(RELEASE_TARGETS)
	@shasum -a 512 $(EXECUTABLES)
	@shasum -a 512 qitmeer-*
checksum: checkversion
	@cat RELEASE.md |grep "$(VERSION) SHA512" -A 8|tail -6|shasum -c
clean:
	@rm -f *.zip
	@rm -f *.tar.gz
	@rm -f ./build/bin/qx
	@rm -f ./build/bin/qitmeer
	@rm -rf ./build/release
