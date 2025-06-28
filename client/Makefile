.PHONY: install dev build preview deploy clean

install:
	pnpm install

dev:
	pnpm run dev  # Hot reload automático com Vite

build:
	pnpm run build

preview:
	pnpm run preview

deploy:
	pnpm run build && pnpm exec gh-pages -d dist

clean:
	rm -rf node_modules dist
