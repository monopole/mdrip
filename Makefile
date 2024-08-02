MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif

# Perform a local build.
# To build a release, use the release target.
$(MYGOBIN)/mdrip:
	releasing/buildWorkspace.sh

# Create a draft release and push it to github.
# Requires go, git, zip, tar, gh (github cli) and env var GH_TOKEN.
# Complains if workspace is dirty, tests fail, tags don't make sense, etc.
.PHONY: release
release: testClean
	(cd releasing; go run . `realpath ..`)

.PHONY: testClean
testClean: clean
	go test ./...

.PHONY: test
test:
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
