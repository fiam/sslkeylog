GOLANG_CI_LINT_VERSION := v2.4.0


.PHONY: lint lint-check lint-fix

# Run linter with autofix
lint: lint-fix

# Run linter in check mode (no fixes)
lint-check:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(GOLANG_CI_LINT_VERSION) golangci-lint run

# Run linter with autofix enabled
lint-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(GOLANG_CI_LINT_VERSION) golangci-lint run --fix
