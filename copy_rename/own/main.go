package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func main() {
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	dbName := os.Getenv("DB_NAME")
	rawIntervalMs := os.Getenv("INSERT_INTERVAL_MS")

	if user == "" || pass == "" || host == "" || port == "" || dbName == "" || rawIntervalMs == "" {
		panic(fmt.Errorf("invalid environment variables. expected MySQL's [USER, PASS, HOST, PORT, DB_NAME] and INSERT_INTERVAL_MS: %+v\n", map[string]string{
			"user":          user,
			"pass":          "##masked##",
			"host":          host,
			"port":          port,
			"dbName":        dbName,
			"intervalMsRaw": rawIntervalMs,
		}))
	}

	db, err := sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo&multiStatements=true",
		user, pass, host, port, dbName))
	if err != nil {
		panic(errors.Wrap(err, "master connection failed"))
	}

	v, err := strconv.Atoi(rawIntervalMs)
	if err != nil {
		panic(errors.Wrapf(err, "unexpected insert interval argument: %s", rawIntervalMs))
	}
	insertIntervalMs := v

	done := make(chan struct{})
	t := time.Now()

	go func(done chan struct{}) {
		createTableSql := `CREATE TABLE dummy_new LIKE dummy`
		addColumnSql := `ALTER TABLE dummy_new ADD COLUMN added_col INT`
		fmt.Printf("start create table. sql = `%s`\n", strings.Join([]string{createTableSql, addColumnSql}, ";"))
		db.MustExec(createTableSql)
		db.MustExec(addColumnSql)
		fmt.Printf("\ncomplete create table.\n")

		insertTriggerSql := `
CREATE TRIGGER insert_new AFTER INSERT ON dummy FOR EACH ROW
BEGIN
  REPLACE INTO dummy_new (id, contents) VALUES (NEW.id, NEW.contents);
END`
		fmt.Printf("start create insert trigger. sql = `%s`\n", insertTriggerSql)
		db.MustExec(insertTriggerSql)
		fmt.Printf("\ncomplete create insert trigger.\n")

		updateTriggerSql := `
CREATE TRIGGER update_new AFTER UPDATE ON dummy FOR EACH ROW
BEGIN
  DELETE FROM dummy_new WHERE id = OLD.id;
  REPLACE INTO dummy_new (id, contents) VALUES (NEW.id, NEW.contents);
END`
		fmt.Printf("start create update trigger. sql = `%s`\n", updateTriggerSql)
		db.MustExec(updateTriggerSql)
		fmt.Printf("\ncomplete create update trigger.\n")

		deleteTriggerSql := `
CREATE TRIGGER delete_new AFTER DELETE ON dummy FOR EACH ROW
BEGIN
  DELETE FROM dummy_new WHERE id = OLD.id;
END`
		fmt.Printf("start create delete trigger. sql = `%s`\n", deleteTriggerSql)
		db.MustExec(deleteTriggerSql)
		fmt.Printf("\ncomplete create delete trigger.\n")

		fmt.Printf("start copy old rows to new.\n")
		err := copyOldRows(db)
		if err != nil {
			panic(errors.Wrap(err, "failed to copy old rows"))
		}
		fmt.Printf("\ncomplete copy old rows.\n")

		// 古い行のコピー(INSERT IGNORE INTO)だけやれば良いように見えてしまうのであえてスリープを入れてTRIGGERの存在意義を出す
		time.Sleep(5 * time.Second)

		renameTableSql := `RENAME TABLE dummy TO dummy_old, dummy_new TO dummy`
		fmt.Printf("start rename table. sql = `%s`\n", renameTableSql)
		db.MustExec(renameTableSql)
		fmt.Printf("\ncomplete rename table.\n")

		close(done)
	}(done)

	go func(done chan struct{}) {
		fmt.Println("Start DDL with INSERT ( '>' is INSERTED COUNTER )")
		for {
			time.Sleep(time.Duration(insertIntervalMs) * time.Millisecond)
			select {
			case _, ok := <-done:
				if !ok {
					return
				}
			default:
				go insert(db)
			}
		}
	}(done)
	<-done
	fmt.Printf("spend time %.2f second\n", float64(time.Now().Sub(t)/time.Millisecond)/1000.0)
}

func insert(db *sqlx.DB) {
	fmt.Print(">")
	db.MustExec(`INSERT INTO dummy (contents) VALUES (?)`, uuid.New().String())
}

func copyOldRows(db *sqlx.DB) error {
	const copyRowsN = 1000
	var max int

	err := db.QueryRow(`SELECT MAX(id) FROM dummy`).Scan(&max)
	if err != nil {
		return errors.Wrap(err, "failed to fetch count of dummy")
	}

	for i := 1; i <= max; {
		var next int
		var begin int
		err := db.QueryRow(`SELECT MIN(id), MAX(id) + 1 FROM (SELECT id FROM dummy WHERE id >= ? ORDER BY id LIMIT ?) AS ids`, i, copyRowsN).Scan(&begin, &next)
		if err != nil {
			panic(err)
		}
		db.MustExec(`INSERT IGNORE INTO dummy_new (id, contents) SELECT id, contents FROM dummy WHERE id BETWEEN ? AND ? ORDER BY id LIMIT ?`, begin, max, copyRowsN)
		i = next

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
