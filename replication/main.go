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

var insertDB *sqlx.DB

func main() {
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")
	masterHost := os.Getenv("MASTER_HOST")
	masterPort := os.Getenv("MASTER_PORT")
	slaveHost := os.Getenv("SLAVE_HOST")
	slavePort := os.Getenv("SLAVE_PORT")
	dbName := os.Getenv("DB_NAME")
	rawIntervalMs := os.Getenv("INSERT_INTERVAL_MS")

	if user == "" || pass == "" ||
		masterHost == "" || masterPort == "" ||
		slaveHost == "" || slavePort == "" || dbName == "" || rawIntervalMs == "" {
		panic(fmt.Errorf("invalid environment variables. expected MySQL's [USER, PASS, HOST, PORT, DB_NAME] and INSERT_INTERVAL_MS: %+v\n", map[string]string{
			"user":          user,
			"pass":          "##masked##",
			"master_host":   masterHost,
			"master_port":   masterPort,
			"slave_host":    slaveHost,
			"slave_port":    slavePort,
			"dbName":        dbName,
			"intervalMsRaw": rawIntervalMs,
		}))
	}

	masterDB, err := sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo",
		user, pass, masterHost, masterPort, dbName))
	if err != nil {
		panic(errors.Wrap(err, "master connection failed"))
	}

	slaveDB, err := sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo",
		user, pass, slaveHost, slavePort, dbName))
	if err != nil {
		panic(errors.Wrap(err, "slave connection failed"))
	}

	v, err := strconv.Atoi(rawIntervalMs)
	if err != nil {
		panic(errors.Wrapf(err, "unexpected insert interval argument: %s", rawIntervalMs))
	}
	insertIntervalMs := v

	done := make(chan struct{})
	t := time.Now()

	go func(done chan struct{}) {
		// setupで作成されるAUTO_INCREMENTの値が1000000近くあるので十分に大きな値を設定する
		slaveDB.MustExec(`ALTER TABLE dummy AUTO_INCREMENT=2000000`)

		addColumnSql := `ALTER TABLE dummy ADD COLUMN added_col INT NULL, ALGORITHM=INPLACE, LOCK=NONE`
		fmt.Printf("start add column sql = `%s`\n", addColumnSql)
		slaveDB.MustExec(addColumnSql)
		fmt.Printf("\ncomplete add column.\n")

		// DDLが完了したら接続先を新しいテーブルへ向ける
		// 実際の業務などでデプロイなどで接続情報を更新したタイミングでこれと同じような挙動をするはず
		switchConnection(slaveDB)

		// 切り替え後もしばらく挿入処理が入った前提で10秒間sleepをかけてみる
		time.Sleep(10 * time.Second)

		close(done)
	}(done)

	go func(done chan struct{}) {
		insertDB = masterDB
		fmt.Println("Start DDL with INSERT ( '>' is INSERTED COUNTER )")
		for {
			time.Sleep(time.Duration(insertIntervalMs) * time.Millisecond)
			select {
			case _, ok := <-done:
				if !ok {
					return
				}
			default:
				go insert(insertDB)
			}
		}
	}(done)
	<-done
	fmt.Printf("\nspend time %.2f second\n", float64(time.Now().Sub(t)/time.Millisecond)/1000.0)

	//revertDDL(insertDB)
}

func insert(db *sqlx.DB) {
	fmt.Print(">")
	db.MustExec(`INSERT INTO dummy (contents) VALUES (?)`, uuid.New().String())
}

func switchConnection(db *sqlx.DB) {
	insertDB = db
}

func revertDDL(db *sqlx.DB) {
	dropColumnSql := `ALTER TABLE dummy DROP COLUMN added_col, ALGORITHM=INPLACE, LOCK=NONE`
	fmt.Printf("start drop (revert add column) column sql = `%s`\n", dropColumnSql)
	db.MustExec(dropColumnSql)
	fmt.Printf("\ncomplete drop column.\n")
}
