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
var twischedules twitter.Schedules

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

var fetchtime time.Time

func _main(ctx context.Context) (string, error) {
	buf, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		log.Fatal("Error while load config : ", err)
	}
	ctx = context.WithValue(ctx, config, buf)

	YudenBot(ctx, []Executor{
		Executor{
			Name: "updater",
			Fnc: func(ctx context.Context) (err error) {
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
				twischedules = s
				return err
			},
			Tick: 30 * time.Minute,
			// Tick: 1 * time.Minute,
		},
		Executor{
			Name: "fetcher",
			Fnc: func(ctx context.Context) (err error) {
				_, err = GetConfig(ctx)
				if err != nil {
					return err
				}
				now := time.Now()
				auth := twitter.GetToken("./.token.yml")
				for _, t := range twischedules {
					if t.Time.After(fetchtime) && t.Time.Before(now) && !t.Executed {
						log.Printf("tweet : %v", t.Message)
						twitter.Tweet(t.Message, auth)
						t.Executed = true
					}
				}
				fetchtime = now
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

	// Schedule(ctx, execList)
	log.Println("Yuden-Bot End.")
}
