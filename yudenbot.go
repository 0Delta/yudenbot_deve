/*
YudenBot is supporter of infra-workshop(インフラ勉強会).

What is infra-workshop(インフラ勉強会) ?

Infra-workshop is japanese online community for study computer infrastructure.
(infra-workshop writes as "インフラ勉強会" in Japanese.)

More information

https://wp.infra-workshop.tech/ (Japanese/日本語)
*/
package main

// とりあえずローカルで動くように

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type ctxkey int

const (
	config ctxkey = iota
)

func main() {
	_main(context.TODO())
}

var events []EventData
var jst, _ = time.LoadLocation("Asia/Tokyo")

func getToken() *TwitterAuth {
	buf, err := ioutil.ReadFile("./.token.yml")
	if err != nil {
		log.Fatal("Error while load token : ", err)
	}
	var auth TwitterAuth
	err = yaml.Unmarshal(buf, &auth)
	if err != nil {
		log.Fatal("Error while unmarshal token: ", err)
	}
	return &auth
}

func _main(ctx context.Context) (string, error) {
	buf, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		log.Fatal("Error while load config : ", err)
	}
	ctx = context.WithValue(ctx, config, buf)
	fetchtime := time.Now()

	YudenBot(ctx, []Executor{
		Executor{
			Name: "updater",
			Fnc: func(ctx context.Context) (err error) {
				events, err = GetEventsFromWordpress("wp.infra-workshop.tech")
				return err
			},
			Tick: 30 * time.Minute,
			// Tick: 1 * time.Minute,
		},
		Executor{
			Name: "fetcher",
			Fnc: func(ctx context.Context) (err error) {
				for _, e := range events {
					t := time.Now()
					d := e.StartDate
					if fetchtime.Before(d) && t.After(d) {
						msg := "-- This is test post --\nはじまるよ！\n" + e.Title + "\n" + e.URL + "\n#インフラ勉強会"
						log.Println("post tweet : \n" + msg)
						tweet(msg, getToken())
					}
					d = e.StartDate.Add(-30 * time.Minute)
					if fetchtime.Before(d) && t.After(d) {
						msg := "-- This is test post --\nもうすぐ始まるよ！\n" + e.Title + "\n" + e.URL + "\n#インフラ勉強会"
						log.Println("post tweet : \n" + msg)
						tweet(msg, getToken())
					}
					d = time.Now()
					d = time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, jst)
					if fetchtime.Before(d) && t.After(d) {
						msg := strings.Join([]string{
							"今日(", t.In(jst).Format("01/02"), ")の #インフラ勉強会 は...\n",
							e.Title, "\n",
							e.StartDate.In(jst).Format("15:04"), " - ", e.EndDate.In(jst).Format("15:04"), "\n",
							e.URL,
						}, "")
						log.Println("post tweet : \n" + msg)
						tweet(msg, getToken())
					}
					fetchtime = t
				}
				return nil
			},
			Tick: 1 * time.Minute,
			// Tick: 1 * time.Minute,
		},
	})
	return fmt.Sprintf("Hello ƛ!"), nil
}

func YudenBot(ctx context.Context, execList []Executor) {
	log.Print("run Yuden-Bot")

	// updater
	// Wordpressから情報Get
	// 書き出す
	// 30分ごと程度

	// fetcher
	// 読み出し
	// 時刻チェック
	// execute()
	// 1分毎
	Schedule(ctx, execList)
	log.Println("Yuden-Bot End.")
}

// executer-d
// discordにpost

// executer-t
// twitterにpost
