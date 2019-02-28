CREATE TABLE dummy (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  contents VARCHAR(100) NOT NULL
)
  ENGINE = innodb,
  CHARSET = utf8mb4;

DELIMITER //
CREATE PROCEDURE gen_rows_power_of_2(n INT)
BEGIN
  DECLARE i INT;
  SET i = 0;
  WHILE i < n DO
    INSERT INTO dummy (contents) SELECT contents FROM dummy;
    SET i = i + 1;
  END WHILE;
END;
//
DELIMITER ;

INSERT INTO dummy (contents) VALUES ('6f3fb9328f916cbb7e35eb395612bdb5003a9510bb76fcee87723ffd929322a2');

CALL gen_rows_power_of_2(20); -- 1024 * 1024 = 1M rows

