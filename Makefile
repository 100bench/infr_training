COMPOSE = docker compose

up:
	$(COMPOSE) up -d db redis
	$(COMPOSE) up migrate
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

cache:
	$(COMPOSE) build --no-cache