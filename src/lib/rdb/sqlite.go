package rdb

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/filer"
	"github.com/8245snake/bikeshare_api/src/lib/static"
)

//CreateSQLite DBを作成する
func CreateSQLite(path string) (*sql.DB, error) {
	qry := `
	create table spotinfo(  
		area character (3) not null ,
		spot character (3) not null ,
		time character (20) not null ,
		count character (3) ,
		primary key (area,spot,time) 
	  ) ;
	`
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(qry)
	if err != nil {
		log.Printf("%q: %s\n", err, qry)
		return nil, err
	}

	if err := os.Chmod(path, 0777); err != nil {
		return nil, err
	}
	return db, nil
}

//GetConnectionSQLite 日付を指定してSQLiteのコネクションを取得
func GetConnectionSQLite(t time.Time) (db *sql.DB, err error) {
	filename := t.Format(filer.ModTimeLayout("yyyy-mm-dd")) + ".db"
	path := filepath.Join(static.DirData, filename)

	if filer.CheckFileExist(path) {
		db, err = sql.Open("sqlite3", path)
	} else {
		db, err = CreateSQLite(path)
	}
	return
}
