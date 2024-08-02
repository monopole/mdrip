MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif

# Perform a local build.
# To build a release, use the release target.
$(MYGOBIN)/mdrip:
	releasing/buildWorkspace.sh

.PHONY: release
release: test
	(cd releasing; go run . `realpath ..`)

.PHONY: test
test: clean
	go test ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -f $(MYGOBIN)/mdrip
	rm -f ./internal/webapp/widget/*/widget.html
	go clean
	go clean -testcache
