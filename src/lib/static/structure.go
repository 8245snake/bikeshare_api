package static

/////////////////////////////////////////////////////////////////////////////////////////////////////////
//  構造体
/////////////////////////////////////////////////////////////////////////////////////////////////////////

//JSpotinfo スクレーパーから受け取ったJSONパース用
type JSpotinfo struct {
	Spotinfo []struct {
		Time  string `json:"time"`
		Area  string `json:"area"`
		Spot  string `json:"spot"`
		Count string `json:"count"`
	} `json:"spotinfo"`
}

//JSpotmaster スクレーパーから受け取ったJSONパース用
type JSpotmaster struct {
	Spotmaster []struct {
		Area string `json:"area"`
		Spot string `json:"spot"`
		Name string `json:"name"`
		Lat  string `json:"lat"`
		Lon  string `json:"lon"`
	} `json:"spotmaster"`
}

//JCount JSONマージャリング構造体 JCountsBodyの要素
type JCount struct {
	Count    string `json:"count"`
	Datetime string `json:"datetime"`
	Day      string `json:"day"`
	Hour     string `json:"hour"`
	Minute   string `json:"minute"`
	Month    string `json:"month"`
	Year     string `json:"year"`
}

//JCountsBody JSONマージャリング構造体
type JCountsBody struct {
	Area        string   `json:"area"`
	Spot        string   `json:"spot"`
	Description string   `json:"description"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	Name        string   `json:"name"`
	Counts      []JCount `json:"counts"`
}

//JPlaces JSONマージャリング構造体
type JPlaces struct {
	Area        string `json:"area"`
	Spot        string `json:"spot"`
	Description string `json:"description"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Name        string `json:"name"`
	Recent      Recent `json:"recent"`
}

//JPlacesBody JSONマージャリング構造体
type JPlacesBody struct {
	Num   int       `json:"num"`
	Items []JPlaces `json:"items"`
}

//Recent 最新の台数情報を格納する
type Recent struct {
	Count    string `json:"count"`
	Datetime string `json:"datetime"`
}

//JDistancesBody JSONマージャリング構造体
type JDistancesBody struct {
	Num   int          `json:"num"`
	Items []JDistances `json:"items"`
}

//JDistances JSONマージャリング構造体
type JDistances struct {
	Area        string `json:"area"`
	Spot        string `json:"spot"`
	Description string `json:"description"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Name        string `json:"name"`
	Distance    string `json:"distance"`
	Recent      Recent `json:"recent"`
}

//JAllPlacesBody JSONマージャリング構造体
type JAllPlacesBody struct {
	Num   int `json:"num"`
	Items []struct {
		Area string `json:"area"`
		Spot string `json:"spot"`
		Name string `json:"name"`
	} `json:"items"`
}

//JAllSpotChiled JSONマージャリング構造体
type JAllSpotChiled struct {
	Area string `json:"area"`
	Spot string `json:"spot"`
	Name string `json:"name"`
}

//JConfig JSONマージャリング構造体
type JConfig struct {
	ChannelSecret            string `json:"channel_secret"`
	ClientID                 string `json:"client_id"`
	ClientSecret             string `json:"client_secret"`
	ImgurID                  string `json:"imgur_id"`
	TwitterAccessToken       string `json:"twitter_access_token"`
	TwitterAccessTokenSecret string `json:"twitter_access_token_secret"`
	TwitterConsumerKey       string `json:"twitter_consumer_key"`
	TwitterConsumerKeySecret string `json:"twitter_consumer_key_secret"`
}

//JGraphResponse Graphリクエストの返信用
type JGraphResponse struct {
	Title  string  `json:"title"`
	Width  string  `json:"width"`
	Height string  `json:"height"`
	URL    string  `json:"url"`
	Item   JPlaces `json:"item"`
}

//JUsers ユーザ情報
type JUsers struct {
	Users []JUser `json:"users"`
}

//JUser ユーザ情報
type JUser struct {
	LineID    string   `json:"line_id"`
	SlackID   string   `json:"slack_id"`
	Favorites []string `json:"favorites"`
	Notifies  []string `json:"notifies"`
	Histories []string `json:"histories"`
}

//JServiceStatus システム稼働状況
type JServiceStatus struct {
	Status     StatusMessage `json:"status"`
	Connection StatusMessage `json:"connection"`
	Scraping   StatusMessage `json:"scraping"`
}
