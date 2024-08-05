MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif

# Perform a local build.
# To build an actual release, use the release target.
$(MYGOBIN)/mdrip:
	releasing/buildWorkspace.sh

# Run an end-to-end test.
.PHONY: testE2E
testE2E: $(MYGOBIN)/mdrip
	./releasing/testE2E.sh $(MYGOBIN)/mdrip

# Run unit tests without a clean (allow reliance on the test cache).
.PHONY: testUnit
testUnit:
	go test ./...

# There's a wee bit of code to generate for enums.
.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -f $(MYGOBIN)/mdrip
	rm -f ./internal/webapp/widget/*/widget.html
	go clean
	go clean -testcache

# Force serial execution of dependencies.
# This only really matters in the release target.
.NOTPARALLEL:

# Create a draft release and push it to github.
# Requires go, git, zip, tar, gh (github cli) and env var GH_TOKEN.
# Complains if workspace is dirty, tests fail, tags don't make sense, etc.
.PHONY: release
release: clean testUnit testE2E
	(cd releasing; go run . `realpath ..`)
