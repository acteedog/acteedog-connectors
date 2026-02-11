CONNECTORS := $(wildcard src/*-connector)
GO_PKGS = ./internal/...

.PHONY: test lint update-schema

test:
	@for c in $(CONNECTORS); do \
		echo "==> test $$c"; \
		(cd $$c && go test $(GO_PKGS)); \
	done

lint:
	@for c in $(CONNECTORS); do \
		echo "==> lint $$c"; \
		(cd $$c && golangci-lint run $(GO_PKGS)); \
	done

update-schema:
	@echo "Generating pdk.gen.go from schema..."
	cd src && xtp plugin init --schema-file acteedog-connector-schema.yaml --template Go --path .tmp-connector || \
	    echo "xtp prepare script failed but ignored"
	@echo "Copying generated file to connectors..."
	for dir in $(CONNECTORS); do \
		cp src/.tmp-connector/pdk.gen.go $$dir/pdk.gen.go; \
	done
	@echo "Cleaning up..."
	rm -rf src/.tmp-connector
	@echo "Done."
