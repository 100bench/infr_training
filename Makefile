COMPOSE = docker compose

up:
	$(COMPOSE) up -d db redis kafka
	$(COMPOSE) up migrate migrate-notify
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

cache:
	$(COMPOSE) build --no-cache