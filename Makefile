.PHONY: up
up:
	docker compose up --build --detach --remove-orphans
	@echo "served at http://localhost:8081"

.PHONY: down
down:
	docker compose down --remove-orphans
