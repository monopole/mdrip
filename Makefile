.PHONY: test
test: clean
	go test ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	go clean
	go clean -testcache
	rm -f ./internal/webapp/widget/*/widget.html
