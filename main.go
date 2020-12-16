package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	loadDir      string = "data"
	writeDir     string = "build"
	userURLBase  string = "https://twitter.com/intent/user?user_id="
	tweetURLBase string = "https://twitter.com/twitter/status/"
)

// TweetJSON ツイート情報
type TweetJSON struct {
	TweetID   string `json:"id_str"`
	ReplyToID string `json:"in_reply_to_status_id_str"`
	IsRetweet bool   `json:"retweeted"`
	User      User   `json:"user"`
	RT        RT     `json:"retweeted_status"`
	Text      string `json:"text"`
	Date      string `json:"created_at"`
}

// Fav お気に入り登録したツイートのツイート情報
type Fav struct {
	TweetID   string `json:"id_str"`
	ReplyToID string `json:"in_reply_to_status_id_str"`
	User      User   `json:"user"`
	Text      string `json:"text"`
	Date      string `json:"created_at"`
}

// User ツイートのユーザ情報
type User struct {
	ID         string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	Image      string `json:"profile_image_url_https"`
}

// RT RTしたツイートのツイート情報
type RT struct {
	ID     string `json:"id_str"`
	RTUser RTUser `json:"user"`
	Text   string `json:"text"`
	Date   string `json:"created_at"`
}

// RTUser RTしたツイートのユーザ情報
type RTUser struct {
	UserID     string `json:"id_str"`
	UserName   string `json:"name"`
	ScreenName string `json:"screen_name"`
	Image      string `json:"profile_image_url_https"`
}

// Tweet リスト化したツイート情報
type Tweet struct {
	TweetID    string
	ReplyToID  string
	IsRetweet  bool
	UserID     string
	UserName   string
	ScreenName string
	UserImage  string
	Text       string
	Date       string
	OriginDate string
}

// ViewData テンプレートへ渡すデータ
type ViewData struct {
	Date       time.Time
	CountTweet int
	CountRT    int
	CountFav   int
	Tweets     []*Tweet
	Favs       []*Fav
}

func fileseek(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		// fmt.Println(err.Error())
		// os.Exit(1)
		fmt.Printf("<ERROR> File seek error\n")
		os.Exit(1)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, fileseek(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}
	return paths
}

func checkDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
}

func twitterTimeFormat(datetime string) string {
	t := twitterTimeParse(datetime)
	return t.Format("2006-01-02 15:04:05")
}

func twitterTimeParse(datetime string) time.Time {
	t, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", datetime)
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	t = t.In(jst)
	return t
}

func tweetTextTrim(text string) string {
	trimmedtext := strings.Replace(text, "\n", "<br>", -1)
	return trimmedtext
}

func makeUserURL(id string) string {
	return userURLBase + id
}

func makeTweetURL(id string) string {
	return tweetURLBase + id
}

func extLink(href string, text string) string {
	return "<a href=\"" + href + "\" rel=\"noreferrer noopener\" target=\"_blank\">" + text + "</a>"
}

func main() {
	var (
		d = flag.String("d", loadDir, "Target directory of batch processing")
		f = flag.String("f", "", "Source JSON file name")
	)
	flag.Parse()

	var readFiles []string
	if *f == "" {
		readFiles = fileseek(*d)
	} else {
		readFiles = append(readFiles, *f)
	}
	total := len(readFiles)

	for i, file := range readFiles {
		dirPath := filepath.Dir(file)
		basename := filepath.Base(file[:len(file)-len(filepath.Ext(file))])

		readFile, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		jsonTweets := make([]*TweetJSON, 0)
		err = json.Unmarshal(readFile, &jsonTweets)
		if err != nil {
			fmt.Printf("[%d/%d] <ERROR> JSON (Tweet) parse error: %s\n", i+1, total, file)
			continue
		}

		tweets := make([]*Tweet, 0)
		var countRt int = 0
		for _, post := range jsonTweets {
			tweet := new(Tweet)
			if post.IsRetweet {
				tweet.TweetID = post.RT.ID
				tweet.UserID = post.RT.RTUser.UserID
				tweet.UserName = post.RT.RTUser.UserName
				tweet.ScreenName = post.RT.RTUser.ScreenName
				tweet.UserImage = post.RT.RTUser.Image
				tweet.Text = tweetTextTrim(post.RT.Text)
				tweet.OriginDate = twitterTimeFormat(post.RT.Date)
				countRt = countRt + 1
			} else {
				tweet.TweetID = post.TweetID
				tweet.UserID = post.User.ID
				tweet.UserName = post.User.Name
				tweet.ScreenName = post.User.ScreenName
				tweet.UserImage = post.User.Image
				tweet.Text = tweetTextTrim(post.Text)
			}
			tweet.IsRetweet = post.IsRetweet
			tweet.ReplyToID = post.ReplyToID
			tweet.Date = twitterTimeFormat(post.Date)

			tweets = append(tweets, tweet)
		}

		// ToDo: ふぁぼ取得元データを切り替える
		favs := make([]*Fav, 0)
		err = json.Unmarshal(readFile, &favs)
		if err != nil {
			fmt.Printf("[%d/%d] <ERROR> JSON (Fav) parse error: %s\n", i+1, total, file)
			continue
		}

		viewData := ViewData{
			Date:       twitterTimeParse(jsonTweets[0].Date),
			CountTweet: len(tweets),
			CountRT:    countRt,
			CountFav:   len(favs),
			Tweets:     tweets,
			Favs:       favs,
		}

		templates := []string{"templates/tweets.html"}
		f := template.FuncMap{
			"makeUserURL":  makeUserURL,
			"makeTweetURL": makeTweetURL,
			"extLink":      extLink,
		}
		tpl, _ := template.New(filepath.Base(templates[0])).Funcs(f).ParseFiles(templates...)
		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, viewData)
		if err != nil {
			fmt.Printf("[%d/%d] <ERROR> Template execute error: %s\n", i+1, total, file)
			continue
		}

		dirPath = strings.Replace(dirPath, loadDir, "", 1)
		checkDir(filepath.Join(writeDir, dirPath))
		writePath := filepath.Join(writeDir, dirPath, basename+".html")

		writeFile, err := os.OpenFile(writePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Printf("[%d/%d] <ERROR> File write error: %s\n", i+1, total, file)
			continue
		}
		w := bufio.NewWriter(writeFile)
		w.WriteString(buf.String())
		w.Flush()

		fmt.Printf("[%d/%d] %s -> %s\n", i+1, total, file, writePath)
	}
	fmt.Printf("End.\n")
}
