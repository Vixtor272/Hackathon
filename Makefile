# Farmi MVP — Farmaenlace hackathon
# Go lives in ~/sdk/go1.27.2 and is linked from ~/bin (see README).
export PATH := $(HOME)/bin:$(PATH)
SHELL := /bin/bash

.PHONY: help install backend frontend build run test smoke clean

help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

install: ## install frontend dependencies
	cd frontend && npm install

backend: ## run the Go API only on :8080 (links point to the Vite dev server on :5173)
	cd backend && STATIC_DIR=none WEB_BASE_URL=http://localhost:5173 go run ./cmd/server

frontend: ## run the Svelte dev server on :5173 (proxies /api to :8080)
	cd frontend && npm run dev

build: ## build the Svelte app (frontend/dist) and the Go binary (backend/bin/farmi)
	cd frontend && npm run build
	cd backend && go build -o bin/farmi ./cmd/server

run: build ## single-process demo: Go serves the built app and the API on :8080
	cd backend && ./bin/farmi

test: ## Go tests + frontend type check + vitest
	cd backend && go vet ./... && go test ./...
	cd frontend && npm run check && npm run test

smoke: ## end-to-end smoke test against a freshly started backend (59 checks)
	python3 scripts/e2e_smoke.py

clean: ## remove build artefacts
	rm -rf backend/bin frontend/dist
