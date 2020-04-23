package logger

import (
	"os"

	"github.com/mbndr/logo"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  ファイルI/O関係
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//_log ロガー
var _log *logo.Logger

//InitLogger ロガー初期化
func InitLogger(path string) error {
	// Receiver for the terminal which logs everything
	cliRec := logo.NewReceiver(os.Stderr, "")
	cliRec.Color = true
	cliRec.Level = logo.DEBUG

	// Helper function to get a os.File with the correct options
	logFile, err := logo.Open(path)
	if err != nil {
		return err
	}

	// Receiver for the log file
	// This will log with level INFO (default) and have no colors activated
	// Also the log format is simpler (f.e. ERRO: Message)
	fileRec := logo.NewReceiver(logFile, "")
	fileRec.Format = "%s: %s"
	fileRec.Level = logo.DEBUG

	// Create the logger
	_log = logo.NewLogger(fileRec, cliRec)
	return nil
}

//Debug デバッグログ
func Debug(text ...interface{}) {
	_log.Debug(text...)
}

//Debugf デバッグログフォーマット付き
func Debugf(format string, param ...interface{}) {
	_log.Debugf(format, param...)
}

//Info インフォメーションログ
func Info(text ...interface{}) {
	_log.Info(text...)
}

//Infof インフォメーションログフォーマット付き
func Infof(format string, param ...interface{}) {
	_log.Infof(format, param...)
}
