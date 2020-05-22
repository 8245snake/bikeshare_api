package static

//StatusMessage システム稼働状況
type StatusMessage string

const (
	//StatusOK 正常
	StatusOK StatusMessage = "OK"
	//StatusNG 問題あり
	StatusNG StatusMessage = "NG"
	//StatusNotConnected DBに接続できませんでした
	StatusNotConnected StatusMessage = "DBに接続できませんでした"
	//StatusScrapingError スクレイピングが滞っています
	StatusScrapingError StatusMessage = "スクレイピングが滞っています"
)
