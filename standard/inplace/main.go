package main

import (
	"fmt"
	"os"
	"strconv"
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
	intervalMsRaw := os.Getenv("INSERT_INTERVAL_MS")

	if user == "" || pass == "" || host == "" || port == "" || dbName == "" || intervalMsRaw == "" {
		panic(fmt.Errorf("invalid environment variables. expected MySQL's [USER, PASS, HOST, PORT, DB_NAME] and INSERT_INTERVAL_MS: %+v\n", map[string]string{
			"user":          user,
			"pass":          "##masked##",
			"host":          dbName,
			"port":          port,
			"dbName":        dbName,
			"intervalMsRaw": intervalMsRaw,
		}))
	}

	db, err := sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo",
		user, pass, host, port, dbName))
	if err != nil {
		panic(errors.Wrap(err, "connection failed"))
	}

	v, err := strconv.Atoi(intervalMsRaw)
	if err != nil {
		panic(errors.Wrapf(err, "unexpected insert interval argument: %s", intervalMsRaw))
	}
	insertIntervalMs := v

	done := make(chan struct{})
	t := time.Now()

	go func(done chan struct{}) {
		addIndexSql := `ALTER TABLE dummy ADD INDEX contents_idx(contents), ALGORITHM=INPLACE, LOCK=NONE`
		fmt.Printf("start add index sql = `%s`\n", addIndexSql)
		db.MustExec(addIndexSql)
		fmt.Printf("\ncomplete add index.\n")

		dropIndexSql := `DROP INDEX contents_idx ON dummy LOCK=NONE`
		fmt.Printf("start drop index sql = `%s`\n", dropIndexSql)
		db.MustExec(dropIndexSql)
		fmt.Printf("\ncomplete drop index.\n")

		addColumnSql := `ALTER TABLE dummy ADD COLUMN added_col INT NULL, ALGORITHM=INPLACE, LOCK=NONE`
		fmt.Printf("start add column sql = `%s`\n", addColumnSql)
		db.MustExec(addColumnSql)
		fmt.Printf("\ncomplete add column.\n")

		// insert lock
		// modifyColumnSql := `ALTER TABLE dummy CHANGE COLUMN added_col added_col SMALLINT NULL`
		// fmt.Printf("start modify column sql = `%s`\n", modifyColumnSql)
		// db.MustExec(modifyColumnSql)
		// fmt.Printf("\ncomplete modify column.\n")

		dropColumnSql := `ALTER TABLE dummy DROP COLUMN added_col, ALGORITHM=INPLACE, LOCK=NONE`
		fmt.Printf("start drop column sql = `%s`\n", dropColumnSql)
		db.MustExec(dropColumnSql)
		fmt.Printf("\ncomplete drop column.\n")
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
	fmt.Printf("spend time %d second\n", time.Now().Sub(t)/time.Second)
}

func insert(db *sqlx.DB) {
	fmt.Print(">")
	db.MustExec(`INSERT INTO dummy (contents) VALUES (?)`, uuid.New().String())
}
