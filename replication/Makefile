DC := docker-compose

MYSQL_ROOT_PASSWORD    := root_pass
MYSQL_DATABASE         := slave_promote_pattern
MYSQL_USER             := learning
MYSQL_PASSWORD         := learning_pass
MYSQL_MASTER_HOST_PORT := 33306
MYSQL_SLAVE_HOST_PORT  := 33307
MYSQL_HOST             := 127.0.0.1
INSERT_INTERVAL_MS     := 500

.PHONY: setup
setup: docker-compose.yaml
	-$(DC) rm -sf master slave
	@$(DC) up -d
	$(DC) exec master bash -c 'until mysqladmin ping --silent; do sleep 1; done'
	$(DC) exec slave bash -c 'until mysqladmin ping --silent; do sleep 1; done'
	until bash -c '$(DC) logs slave | grep "replication started in log"'; do sleep 1; done
	: setup done

docker-compose.yaml: docker-compose.yaml.tmpl
	sed -e 's/{{MYSQL_ROOT_PASSWORD}}/$(MYSQL_ROOT_PASSWORD)/' \
	  -e 's/{{MYSQL_DATABASE}}/$(MYSQL_DATABASE)/' \
	  -e 's/{{MYSQL_USER}}/$(MYSQL_USER)/' \
	  -e 's/{{MYSQL_MASTER_HOST_PORT}}/$(MYSQL_MASTER_HOST_PORT)/' \
	  -e 's/{{MYSQL_SLAVE_HOST_PORT}}/$(MYSQL_SLAVE_HOST_PORT)/' \
	  -e 's/{{MYSQL_PASSWORD}}/$(MYSQL_PASSWORD)/' \
	  $< > $@

.PHONY: test_online_ddl
test_online_ddl: main
	: ====================================================
	@USER=$(MYSQL_USER) \
	  PASS=$(MYSQL_PASSWORD) \
	  MASTER_HOST=$(MYSQL_HOST) \
	  MASTER_PORT=$(MYSQL_MASTER_HOST_PORT) \
	  SLAVE_HOST=$(MYSQL_HOST) \
	  SLAVE_PORT=$(MYSQL_SLAVE_HOST_PORT) \
	  DB_NAME=$(MYSQL_DATABASE) \
	  INSERT_INTERVAL_MS=$(INSERT_INTERVAL_MS) \
	./$<
	: ====================================================
main: main.go
	go build main.go

.PHONY: ssh/master ssh/slave
ssh/master:
	$(DC) exec $(@F) '/bin/bash'
ssh/slave:
	$(DC) exec $(@F) '/bin/bash'

.PHONY: cli/master cli/slave
cli/master:
	$(DC) exec master mysql -P $(MYSQL_MASTER_HOST_PORT) -u root -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)
cli/slave:
	$(DC) exec slave mysql -P $(MYSQL_SLAVE_HOST_PORT) -u root -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)

.PHONY: clean
clean:
	-$(DC) rm -sf master slave
	-rm docker-compose.yaml
