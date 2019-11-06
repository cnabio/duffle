COMPOSE=/usr/local/bin/docker-compose

.PHONY: install
install:
	$(COMPOSE) up -d
	$(COMPOSE) ps

.PHONY: upgrade
upgrade: install

.PHONY: uninstall
uninstall:
	$(COMPOSE) down

.PHONY: status
status:
	$(COMPOSE) ps
