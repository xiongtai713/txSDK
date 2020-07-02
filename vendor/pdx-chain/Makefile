# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: utopia

GOBIN = $(shell pwd)/build/bin
GO ?= latest

utopia:
	build/prv/env.sh go run deps.go
	build/env.sh go run build/ci.go install ./cmd/utopia
	@echo "Done building."
	@echo "Run \"$(GOBIN)/utopia\" to launch utopia."

plume:
	build/prv/env.sh go run deps.go
	go build -ldflags "-w -s" -o ./build/bin/plume$(VERSION) ./cmd/plume
	@echo "Done building."

plume-ios:
	gomobile bind -ldflags "-w -s" -target=ios -o=build/bin/plume$(VERSION).framework pdx-chain/cmd/plume/mobile
	tar czvf build/bin/plume$(VERSION).framework.tar.gz build/bin/plume$(VERSION).framework
	@echo "Done building."

plume-android:
	gomobile bind -ldflags "-w -s" -target=android -o=build/bin/plume$(VERSION).aar pdx-chain/cmd/plume/mobile
	@echo "Done building."

clean:
	./build/clean_go_build_cache.sh

deps:
	rm -rf ./build/_deps/*
	build/prv/env.sh go run deps.go
