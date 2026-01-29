COMPOSE = docker compose

up:
	$(COMPOSE) up -d db
	$(COMPOSE) run --rm migrate
	$(COMPOSE) up -d app
	$(COMPOSE) up -d redis

down:
	$(COMPOSE) down

restart:
	make down
	make up

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

migrate:
	$(COMPOSE) run --rm migrate
