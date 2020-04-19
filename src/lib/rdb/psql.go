package rdb

import (
	"database/sql"
	"fmt"

	"github.com/8245snake/bikeshare_api/src/lib/filer"

	_ "github.com/lib/pq"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  変数
/////////////////////////////////////////////////////////////////////////////////////////////////////////
var _host string
var _port string
var _user string
var _password string
var _dbname string

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  関数
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//GetConnectionPsql DB接続
func GetConnectionPsql() (*sql.DB, error) {
	//初めて呼ばれたときのみ設定読み込みをする
	if _host == "" {
		if err := initConnectionSetting(); err != nil {
			return nil, err
		}
	}
	connectstring := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", _host, _port, _user, _password, _dbname)
	return sql.Open("postgres", connectstring)
}

//initConnectionSetting 各種ディレクトリ情報を変数に格納する
func initConnectionSetting() error {
	section := "DB"
	_host = filer.GetIniData(section, "HOST", "")
	_port = filer.GetIniData(section, "PORT", "")
	_user = filer.GetIniData(section, "USER", "")
	_password = filer.GetIniData(section, "PASSWORD", "")
	_dbname = filer.GetIniData(section, "DB_NAME", "")
	if _host == "" || _port == "" || _user == "" || _password == "" || _dbname == "" {
		return fmt.Errorf("DB config is wrong")
	}
	return nil
}
