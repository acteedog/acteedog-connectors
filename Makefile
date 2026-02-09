CONNECTORS := github-connector slack-connector
GO_PKGS = ./internal/...

.PHONY: test lint

test:
	@for c in $(CONNECTORS); do \
		echo "==> test $$c"; \
		(cd src/$$c && go test $(GO_PKGS)); \
	done

lint:
	@for c in $(CONNECTORS); do \
		echo "==> lint $$c"; \
		(cd src/$$c && golangci-lint run $(GO_PKGS)); \
	done
