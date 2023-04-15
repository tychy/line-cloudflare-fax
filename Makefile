
.PHONY: build
build:
	mkdir -p dist
	GOOS=js GOARCH=wasm go build -o ./dist/app.wasm ./...

update-workers:
	cd workers && git pull
.PHONY: dev
dev:
	wrangler dev

.PHONY: prod
prod:
	wrangler publish -e production
