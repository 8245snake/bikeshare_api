package filer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8245snake/bikeshare_api/src/lib/logger"
	"github.com/8245snake/bikeshare_api/src/lib/static"

	"gopkg.in/ini.v1"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  ファイルI/O関係
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//_ini iniオブジェクト
var _ini *ini.File

//GetIniData 設定ファイル読み込み
func GetIniData(section string, key string, defaultval string) string {
	if _ini == nil {
		return defaultval
	}
	return _ini.Section(section).Key(key).MustString(defaultval)
}

//GetIniDataInt 設定ファイル読み込み（数値）
func GetIniDataInt(section string, key string, defaultval int) int {
	if _ini == nil {
		return defaultval
	}
	return _ini.Section(section).Key(key).MustInt(defaultval)
}

//CheckFileExist ファイルがあるかチェックする。ない場合はメッセージ出力しFalseを返す。
func CheckFileExist(path string) bool {
	if f, err := os.Stat(path); os.IsNotExist(err) || f.IsDir() {
		fmt.Println(fmt.Sprintf("File '%s' is not exist", filepath.Clean(path)))
		return false
	}
	return true
}

//CheckDirectoryExist ディレクトリがあるかチェックする。ない場合はメッセージ出力しFalseを返す。
func CheckDirectoryExist(path string) bool {
	if f, err := os.Stat(path); os.IsNotExist(err) || !f.IsDir() {
		fmt.Println(fmt.Sprintf("Derectory '%s' is not exist", filepath.Clean(path)))
		return false
	}
	return true
}

//FileCopy ファイルをコピーする
func FileCopy(srcPath string, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

//FileMove 指定したフォルダにファイルを移動する
func FileMove(srcPath string, dstDir string) error {
	_, fileName := filepath.Split(srcPath)
	_ = os.MkdirAll(dstDir, 0777)
	dstPath := filepath.Clean(filepath.Join(dstDir, fileName))
	if err := FileCopy(srcPath, dstPath); err != nil {
		return err
	} else {
		os.Remove(srcPath)
	}
	return nil
}

//SetCurDirToExeDir カレントパスをexeのパスにする
func SetCurDirToExeDir() {
	exe, _ := os.Executable()
	static.DirExe = filepath.Clean(filepath.Dir(exe))
	fmt.Println(static.DirExe)
	os.Chdir(static.DirExe)
}

//InitDirSetting 各種ディレクトリ情報を変数に格納する
func InitDirSetting() error {
	//カレントパス設定
	SetCurDirToExeDir()
	//iniの存在チェック
	if !CheckFileExist(static.IniPath) {
		return fmt.Errorf("app.ini is not exist")
	}
	//iniをキャッシュ
	if c, err := ini.Load(static.IniPath); err == nil {
		_ini = c
	} else {
		_ini = nil
		return err
	}

	//ロガーを初期化
	if err := logger.InitLogger(filepath.Join(static.DirLog, GetExeName()+".log")); err != nil {
		return err
	}

	return nil
}

//ModTimeLayout 時刻フォーマットを分かりやすい形式から変換
func ModTimeLayout(layout string) (newLayout string) {
	newLayout = layout
	newLayout = strings.Replace(newLayout, "yyyy", "2006", -1)
	newLayout = strings.Replace(newLayout, "mm", "01", -1)
	newLayout = strings.Replace(newLayout, "dd", "02", -1)
	newLayout = strings.Replace(newLayout, "HH", "15", -1)
	newLayout = strings.Replace(newLayout, "hh", "03", -1)
	newLayout = strings.Replace(newLayout, "MM", "04", -1)
	newLayout = strings.Replace(newLayout, "SS", "05", -1)
	return
}

//GetFileNameWithoutExt ファイルパスから拡張子を除いたファイル名を返す
func GetFileNameWithoutExt(path string) string {
	// Fixed with a nice method given by mattn-san
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

//GetExeName 実行ファイル名から拡張子を除いた文字列を返す
func GetExeName() string {
	return GetFileNameWithoutExt(os.Args[0])
}

//WaitForFileCreation ファイルができるまで待つ
//監視間隔とタイムアウトを秒で指定
//見つかったらtrueを返す
func WaitForFileCreation(path string, watchInterval float32, timeout float32) (exists bool) {
	exists = false
	tryNum := int(timeout/watchInterval) + 1
	for i := 0; i < tryNum; i++ {
		if CheckFileExist(path) {
			exists = true
			break
		}
		time.Sleep(time.Duration(watchInterval) * time.Second)
	}
	return
}
