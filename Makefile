
EXECUTABLE := qitmeer
GITVER := $(shell git rev-parse --short=7 HEAD )
GITDIRTY := $(shell git diff --quiet || echo '-dirty')
GITVERSION = "$(GITVER)$(GITDIRTY)"
DEV=dev
RELEASE=release
LDFLAG_DEV = -X github.com/Qitmeer/qitmeer/version.Build=$(DEV)-$(GITVERSION)
LDFLAG_RELEASE = -X github.com/Qitmeer/qitmeer/version.Build=$(RELEASE)-$(GITVERSION)
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

DEV_EXECUTABLES := \
	build/dev/darwin/amd64/bin/$(EXECUTABLE) \
	build/dev/linux/amd64/bin/$(EXECUTABLE) \
	build/dev/windows/amd64/bin/$(EXECUTABLE).exe

COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.tar.gz) $(WIN_EXECUTABLES:%.exe=%.zip) $(WIN_EXECUTABLES:%.exe=%.cn.zip)

RELEASE_TARGETS=$(EXECUTABLES) $(COMPRESSED_EXECUTABLES)

DEV_TARGETS=$(DEV_EXECUTABLES)

ZMQ = FALSE

.PHONY: qitmeer qx release

qitmeer: qitmeer-build
	@echo "Done building."
	@echo "  $(shell $(GOBIN)/qitmeer --version))"
	@echo "Run \"$(GOBIN)/qitmeer\" to launch."

qitmeer-build:
    ifeq ($(ZMQ),TRUE)
		@echo "Enalbe ZMQ"
		@go build -o $(GOBIN)/qitmeer $(GOFLAGS_DEV) -tags=zmq "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
    else
		@go build -o $(GOBIN)/qitmeer $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
    endif
qx:
	@go build -o $(GOBIN)/qx $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/qx"
burn:
	@go build -o $(GOBIN)/burn $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/burn"
relay:
	@go build -o $(GOBIN)/relaynode $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/relaynode"
fastibd:
	@go build -o $(GOBIN)/fastibd $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/fastibd"


checkversion: qitmeer-build
#	@echo version $(VERSION)

all: qitmeer-build qx burn relay fastibd

# amd64 release
build/release/%: OS=$(word 3,$(subst /, ,$(@)))
build/release/%: ARCH=$(word 4,$(subst /, ,$(@)))
build/release/%/$(EXECUTABLE):
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
build/release/%/$(EXECUTABLE).exe:
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"

# amd64 dev
build/dev/%: OS=$(word 3,$(subst /, ,$(@)))
build/dev/%: ARCH=$(word 4,$(subst /, ,$(@)))
build/dev/%/$(EXECUTABLE):
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_DEV) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"
build/dev/%/$(EXECUTABLE).exe:
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_DEV) -o $(@) "github.com/Qitmeer/qitmeer/cmd/qitmeerd"


%.zip: %.exe
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).zip "$<"

%.cn.zip: %.exe
	@echo Build $(@).cn.zip
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@zip -j $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).cn.zip "$<" script/win/start.bat

%.tar.gz : %
	@echo tar $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@tar -zcvf $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).tar.gz "$<"
release: clean checkversion
	@echo "Build release version : $(VERSION)"
	@$(MAKE) $(RELEASE_TARGETS)
	@shasum -a 512 $(EXECUTABLES) > $(EXECUTABLE)-$(VERSION)_checksum.txt
	@shasum -a 512 $(EXECUTABLE)-$(VERSION)-* >> $(EXECUTABLE)-$(VERSION)_checksum.txt
dev: clean checkversion
	@echo "Build dev version : $(VERSION)"
	@$(MAKE) $(DEV_TARGETS)

checksum: checkversion
	@cat $(EXECUTABLE)-$(VERSION)_checksum.txt|shasum -c
clean:
	@rm -f *.zip
	@rm -f *.tar.gz
	@rm -f ./build/bin/qx
	@rm -f ./build/bin/qitmeer
	@rm -rf ./build/release
	@rm -rf ./build/dev
