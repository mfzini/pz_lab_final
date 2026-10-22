.PHONY: back front

back:
	cd backend && \
	go run cmd/server/main.go

# Los scripts de package.json invocan las herramientas con `bun <archivo>`
# explícito: la imagen del devcontainer no trae node, y si no, los binarios
# (vite/eslint/oxlint) mueren con `env: 'node': No such file or directory`.
front:
	cd frontend && \
	bun run dev
