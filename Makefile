COMPOSE = docker compose

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

cache:
	$(COMPOSE) build --no-cache