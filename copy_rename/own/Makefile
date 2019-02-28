DC := docker-compose

MYSQL_ROOT_PASSWORD := root_pass
MYSQL_DATABASE      := rename_pattern
MYSQL_USER          := learning
MYSQL_PASSWORD      := learning_pass
MYSQL_HOST_PORT     := 33306
MYSQL_HOST          := 127.0.0.1
INSERT_INTERVAL_MS  := 500

CONTAINER_NAME := mysql_learn_own

.PHONY: setup
setup:
	-docker rm -f $(CONTAINER_NAME)
	@docker run --rm --name $(CONTAINER_NAME) -itd \
	  -p $(MYSQL_HOST_PORT):3306 \
	  -v '$(CURDIR)/initdb.d:/docker-entrypoint-initdb.d:ro' \
	  -e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
      -e MYSQL_DATABASE=$(MYSQL_DATABASE) \
      -e MYSQL_USER=$(MYSQL_USER) \
      -e MYSQL_PASSWORD=$(MYSQL_PASSWORD) \
	  mysql:5.6
	docker exec $(CONTAINER_NAME) bash -c 'until mysqladmin ping --silent; do sleep 1; done'
	: setup done

.PHONY: test_online_ddl
test_online_ddl: main
	: ====================================================
	@USER=$(MYSQL_USER) \
	  PASS=$(MYSQL_PASSWORD) \
	  HOST=$(MYSQL_HOST) \
	  PORT=$(MYSQL_HOST_PORT) \
	  DB_NAME=$(MYSQL_DATABASE) \
	  INSERT_INTERVAL_MS=$(INSERT_INTERVAL_MS) \
	./$<
	: ====================================================
main: main.go
	go build main.go

.PHONY: ssh
ssh:
	docker exec -it $(CONTAINER_NAME) '/bin/bash'

.PHONY: cli
cli:
	@docker exec -it $(CONTAINER_NAME) mysql -P $(MYSQL_HOST_PORT) -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

.PHONY: clean
clean:
	-docker rm -f $(CONTAINER_NAME)
