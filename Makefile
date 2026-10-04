back:
	cd backend && \
	go run cmd/server/main.go
front:
	cd frontend &&\
	bun dev