# Farmi MVP — Farmaenlace hackathon
# Go lives in ~/sdk/go1.27.2 and is linked from ~/bin (see README).
export PATH := $(HOME)/bin:$(PATH)
SHELL := /bin/bash

.PHONY: help install backend frontend cliente empresa build run test smoke clean

help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

install: ## install frontend dependencies
	cd frontend && npm install

backend: ## run the Go API only: customer surface :8080, company surface :8081 (links point to Vite :5173)
	cd backend && STATIC_DIR=none COMPANY_STATIC_DIR=none WEB_BASE_URL=http://localhost:5173 go run ./cmd/server

frontend: ## run both Svelte dev servers: cliente :5173 (→ :8080) and empresa :5174 (→ :8081)
	cd frontend && trap 'kill 0' EXIT INT TERM; npm run dev & npm run dev:empresa & wait

cliente: ## run only the customer app dev server on :5173 (WhatsApp, checkout, DeUna)
	cd frontend && npm run dev

empresa: ## run only the company portal dev server on :5174 (caja, reparto, notificaciones)
	cd frontend && npm run dev:empresa

build: ## build both Svelte apps (frontend/dist/{cliente,empresa}) and the Go binary (backend/bin/farmi)
	cd frontend && npm run build
	cd backend && go build -o bin/farmi ./cmd/server

run: build ## single-process demo: cliente app + API on :8080, empresa portal + API on :8081
	cd backend && ./bin/farmi

test: ## Go tests + frontend type check + vitest
	cd backend && go vet ./... && go test ./...
	cd frontend && npm run check && npm run test

smoke: ## end-to-end smoke test against a freshly started backend (66 checks, both surfaces)
	python3 scripts/e2e_smoke.py

clean: ## remove build artefacts
	rm -rf backend/bin frontend/dist
