CHANGE MASTER TO MASTER_HOST='master', MASTER_USER='repl', MASTER_AUTO_POSITION = 1;
START SLAVE;