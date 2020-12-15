package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

type TweetJSON struct {
	TweetID   string `json:"id_str"`
	ReplyToID string `json:"in_reply_to_status_id_str"`
	IsRetweet bool   `json:"retweeted"`
	User      User   `json:"user"`
	RT        RT     `json:"retweeted_status"`
	Text      string `json:"text"`
	Date      string `json:"created_at"`
}

type User struct {
	ID         string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type RT struct {
	ID     string `json:"id_str"`
	RTUser RTUser `json:"user"`
	Text   string `"json:text"`
}

type RTUser struct {
	UserID     string `json:"id_str"`
	UserName   string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type Tweet struct {
	TweetID    string
	ReplyToID  string
	IsRetweet  bool
	UserID     string
	UserName   string
	ScreenName string
	Text       string
	Date       string
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("templates/tweets.html"))
}

func main() {
	raw, err := ioutil.ReadFile("data/test_large.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	jsonTweets := make([]*TweetJSON, 0)
	err = json.Unmarshal(raw, &jsonTweets)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tweets := make([]*Tweet, 0)
	for _, post := range jsonTweets {
		tweet := new(Tweet)
		if post.IsRetweet {
			tweet.TweetID = post.RT.ID
			tweet.UserID = post.RT.RTUser.UserID
			tweet.UserName = post.RT.RTUser.UserName
			tweet.ScreenName = post.RT.RTUser.ScreenName
			tweet.Text = post.RT.Text
		} else {
			tweet.TweetID = post.TweetID
			tweet.UserID = post.User.ID
			tweet.UserName = post.User.Name
			tweet.ScreenName = post.User.ScreenName
			tweet.Text = post.Text
		}
		tweet.IsRetweet = post.IsRetweet
		tweet.ReplyToID = post.ReplyToID
		tweet.Date = post.Date

		tweets = append(tweets, tweet)
	}

	err = tpl.Execute(os.Stdout, tweets)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
