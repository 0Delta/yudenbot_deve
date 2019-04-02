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
	"sync"
	"time"

	"github.com/0Delta/colog2slack"
	"github.com/0Delta/yudenbot_devel/discord"
	"github.com/0Delta/yudenbot_devel/eventdata"
	"github.com/0Delta/yudenbot_devel/twitter"
	yaml "gopkg.in/yaml.v2"
)

type ctxkey int

const (
	config ctxkey = iota
)

func main() {
	_main(context.Background())
}

var events []eventdata.EventData
var jst, _ = time.LoadLocation("Asia/Tokyo")

var mtscs sync.Mutex
var twischedules twitter.Schedules

type discordschedule struct {
	Event    eventdata.EventData
	Time     time.Time
	Executed bool
}

var mdscs sync.Mutex
var discordschedules []discordschedule

type configArgs struct {
	WordpressURL    string `yaml:"wordpressurl"`
	DayLine         int    `yaml:"dayline"`
	NextPreviewHour int    `yaml:"nextpreviewhour"`
	SummaryPostHour int    `yaml:"summaryposthour"`
}

func GetConfig(ctx context.Context) (args *configArgs, err error) {
	v := ctx.Value(config)
	buf, ok := v.([]byte)
	if !ok {
		log.Fatal("Error while load token : ", fmt.Errorf("token not found"))
		return nil, err
	}
	err = yaml.Unmarshal(buf, &args)
	if err != nil {
		log.Fatal("Error while unmarshal token: ", err)
		return nil, err
	}
	return args, nil
}

type secretConf struct {
	SlackURL4Log string `yaml:"slackurlforlog" json:"slackurlforlog"`
}

func GetToken(filepath string) *secretConf {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Error while load token : ", err)
	}
	var conf secretConf
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatal("Error while unmarshal token: ", err)
	}
	return &conf
}

var fetchtime time.Time

func _main(ctx context.Context) (string, error) {
	// logfile, err := os.OpenFile("./test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	panic("cannnot open test.log:" + err.Error())
	// }
	// defer logfile.Close()
	// log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	buf, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		log.Fatal("Error while load config : ", err)
	}
	ctx = context.WithValue(ctx, config, buf)
	conf := GetToken(".token.yml")
	colog2slack.Enable(conf.SlackURL4Log)

	YudenBot(ctx, []Executor{
		Executor{
			Name: "updater",
			Fnc:  updater,
			Tick: 30 * time.Minute,
			// Tick: 1 * time.Minute,
		},
		Executor{
			Name: "fetcher",
			Fnc:  fetcher,
			Tick: 1 * time.Minute,
			// Tick: 1 * time.Minute,
		},
		Executor{
			Name: "discord",
			Fnc:  createAndPostDiscordChannel,
			Tick: 1 * time.Minute,
		},
	})
	return fmt.Sprintf("Hello ƛ!"), nil
}

func UpdateTwitterScedules(newsc twitter.Schedules) {
	mtscs.Lock()
	defer mtscs.Unlock()
	twischedules = newsc
}

func UpdateDiscordScedules(newsc []discordschedule) {
	mdscs.Lock()
	defer mdscs.Unlock()
	discordschedules = newsc
}

func YudenBot(ctx context.Context, execList []Executor) {
	log.Print("run Yuden-Bot")

	Schedule(ctx, execList)
	log.Println("Yuden-Bot End.")
}

// compornents
func updater(ctx context.Context) (err error) {
	conf, err := GetConfig(ctx)
	if err != nil {
		return err
	}
	events, err = eventdata.GetEventsFromWordpress(conf.WordpressURL, conf.DayLine)
	if err != nil {
		return err
	}
	// update tweetSchedule
	d := time.Now()
	dayLine := time.Date(d.Year(), d.Month(), d.Day(), conf.DayLine, 0, 0, 0, jst).Add(24 * time.Hour)
	nextPostHour := time.Date(d.Year(), d.Month(), d.Day(), conf.SummaryPostHour, 0, 0, 0, jst)
	for _, e := range events {
		if e.EndDate.After(nextPostHour) && e.StartDate.Before(dayLine) {
			nextPostHour = e.EndDate
		}
	}
	var s twitter.Schedules
	var disSc []discordschedule
	for _, e := range events {
		// start
		s.Append(e,
			e.StartDate,
			strings.Join([]string{
				"はじまるよ！", "\n",
				e.Title, "\n",
				e.URL, "\n",
				"#インフラ勉強会",
			}, ""),
		)
		// remind
		s.Append(e,
			e.StartDate.Add(-30*time.Minute),
			strings.Join([]string{
				"もうすぐ始まるよ！\n", e.Title, "\n",
				e.URL, "\n",
				"#インフラ勉強会",
			}, ""),
		)
		disSc = append(disSc, discordschedule{e, e.StartDate.Add(-30 * time.Minute), false})
		// today's summary
		d = time.Now()
		if e.StartDate.Before(dayLine) {
			s.Append(e,
				time.Date(d.Year(), d.Month(), d.Day(), conf.SummaryPostHour, 0, 0, 0, jst),
				strings.Join([]string{
					"今日(", d.In(jst).Format("01/02"), ")の #インフラ勉強会 は...\n",
					e.Title, "\n",
					e.StartDate.In(jst).Format("15:04"), " - ", e.EndDate.In(jst).Format("15:04"), "\n",
					e.URL,
				}, ""),
			)
		}
		// next
		d = d.Add(24 * time.Hour)
		if e.StartDate.After(dayLine) && e.StartDate.Before(dayLine.Add(24*time.Hour)) {
			s.Append(e,
				nextPostHour,
				strings.Join([]string{
					"#インフラ勉強会 、次回(", d.In(jst).Format("01/02"), ")は...\n",
					e.Title, "\n",
					e.StartDate.In(jst).Format("15:04"), " - ", e.EndDate.In(jst).Format("15:04"), "\n",
					e.URL,
				}, ""),
			)
		}
	}
	UpdateTwitterScedules(s)
	UpdateDiscordScedules(disSc)
	return err
}

func fetcher(ctx context.Context) (err error) {
	_, err = GetConfig(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	// auth := twitter.GetToken("./.token.yml")
	for _, t := range twischedules {
		if t.Time.After(fetchtime) && t.Time.Before(now) && !t.Executed {
			log.Printf("tweet : %v", t.Message)
			// twitter.Tweet(t.Message, auth)
			t.Executed = true
		}
	}
	fetchtime = now
	return nil
}

func createAndPostDiscordChannel(ctx context.Context) (err error) {
	_, err = GetConfig(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	auth := discord.GetToken("./.token.yml")
	for _, d := range discordschedules {
		if d.Time.After(fetchtime) && d.Time.Before(now) && !d.Executed {
			log.Printf("discord : %v", d.Event.Title)
			s := discord.GetDiscord(auth.Token)
			chname := fmt.Sprintf("%s-%s", d.Event.StartDate.Format("0102"), d.Event.Title)
			c := discord.CreateTextChannel(s, auth.GuildID, chname)

			// create post message
			message := fmt.Sprintln(d.Event.Title)
			message += fmt.Sprintln(d.Event.URL)
			message += fmt.Sprintf("%s 〜 %s", d.Event.StartDate.Format("01:02"), d.Event.StartDate.Format("01:02"))
			message += fmt.Sprintln()
			message += fmt.Sprintln("----")
			message += d.Event.Description
			discord.SendMessage(s, c, message)
			message = `■ ご注意!!
音声チャンネルは Study-Group01 です。入室時には意図せずマイクがオンのままになっていないかご確認をお願いします。
http://bit.ly/2HWB9ZL
進行に影響がある場合は一旦 AFK 部屋に移動させて頂く場合がありますのでその際はマイクをミュートにしつつ戻ってきてくだされば。

■ 質問したいとき
頭に "Q. " をつけてコメントしておいてくだされば。あとで主催者が拾います。

■ 匿名で質問したいとき
質問箱 BOT さんに "Q. " の付いた質問を投げるとチャンネルに匿名で投稿し直してくれます。
勉強会中、みんなの前だとちょっと質問しづらいな‥って思ったら質問箱 BOT に "Q." が先頭についたメッセージを送ってください。
http://bit.ly/2rjZyjL`
			discord.SendMessage(s, c, message)
			d.Executed = true
		}
	}
	fetchtime = now
	return nil
}
