EXECUTABLE := qitmeer-miner
GITVERSION := $(shell git rev-parse --short HEAD)
DEV=dev
RELEASE=release
LDFLAG_DEV = -X github.com/Qitmeer/github.com/Qitmeer/qitmeer/cmd/miner/version.Build=$(DEV)-$(GITVERSION)
LDFLAG_RELEASE = -X github.com/Qitmeer/github.com/Qitmeer/qitmeer/cmd/miner/version.Build=$(RELEASE)-$(GITVERSION)
GOFLAGS_DEV = -ldflags "$(LDFLAG_DEV)"
TAGS=cpu
GOFLAGS_RELEASE = -ldflags "$(LDFLAG_RELEASE)"
VERSION=$(shell ./build/bin/qitmeer-miner --version | grep ^qitmeer-miner | cut -d' ' -f3|cut -d'+' -f1)
GOBIN = ./build/bin

UNIX_EXECUTABLES := \
	build/release/darwin/amd64/bin/$(EXECUTABLE) \
	build/release/linux/amd64/bin/$(EXECUTABLE)
WIN_EXECUTABLES := \
	build/release/windows/amd64/bin/$(EXECUTABLE).exe

EXECUTABLES=$(UNIX_EXECUTABLES) $(WIN_EXECUTABLES)
	
COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.tar.gz) $(WIN_EXECUTABLES:%.exe=%.zip)

RELEASE_TARGETS=$(EXECUTABLES) $(COMPRESSED_EXECUTABLES)

.PHONY: qitmeer-miner release

qitmeer-miner: miner-build
	@echo "Done building."
	@echo "  $(shell $(GOBIN)/qitmeer-miner --version))"
	@echo "Run \"$(GOBIN)/qitmeer-miner\" to launch."

miner-build:
	@go build -o $(GOBIN)/qitmeer-miner  -tags $(TAGS) $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer/cmd/miner"
checkversion: miner-build
#	@echo version $(VERSION)

all: qitmeer-miner

# amd64 release
build/release/%: OS=$(word 3,$(subst /, ,$(@)))
build/release/%: ARCH=$(word 4,$(subst /, ,$(@)))
build/release/%/$(EXECUTABLE):
	@echo Build $(@)
	go build $(GOFLAGS_RELEASE) -o $(@) -tags $(TAGS) "github.com/Qitmeer/qitmeer/cmd/miner"
build/release/%/$(EXECUTABLE).exe:
	@echo Build $(@)
	go build $(GOFLAGS_RELEASE) -o $(@) -tags $(TAGS) "github.com/Qitmeer/qitmeer/cmd/miner"
%.zip: %.exe
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@cp example.solo.conf solo.conf
	@cp example.pool.conf pool.conf
	@sh installLibrary.sh
	@zip -j $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).zip "$<" libcudacuckoo.so libcuckoo.so cuckoo.dll cudacuckoo.dll libcuckoo.dylib mining.bat solo.conf pool.conf

%.tar.gz : %
	@echo tar $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@tar -zcvf $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).tar.gz "$<"
release: clean checkversion
	@echo "Build release version : $(VERSION)"
	@$(MAKE) $(RELEASE_TARGETS)
	@shasum -a 512 $(EXECUTABLES) > $(EXECUTABLE)-$(VERSION)_checksum.txt
	@shasum -a 512 $(EXECUTABLE)-$(VERSION)-* >> $(EXECUTABLE)-$(VERSION)_checksum.txt
checksum: checkversion
	@cat $(EXECUTABLE)-$(VERSION)_checksum.txt|shasum -c
clean:
	@rm -f *.zip
	@rm -f *.tar.gz
	@rm -f ./build/bin/qitmeer-miner
	@rm -rf ./build/release
