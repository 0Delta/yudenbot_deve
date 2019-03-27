package twitter

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/0Delta/yudenbot_devel/eventdata"
	"github.com/ChimeraCoder/anaconda"
	"gopkg.in/yaml.v2"
)

var twitterAPI *anaconda.TwitterApi
var jst, _ = time.LoadLocation("Asia/Tokyo")
var apiHash []byte

type TwitterAuth struct {
	ConsumerKey    string `yaml:"consumerKey" json:"consumerKey"`
	ConsumerSecret string `yaml:"consumerSecret" json:"consumerSecret"`
	AccessToken    string `yaml:"accessToken" json:"accessToken"`
	AccessSecret   string `yaml:"accessSecret" json:"accessSecret"`
}

func GetToken(filepath string) *TwitterAuth {
	buf, err := ioutil.ReadFile(filepath)
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

func getTwitterAPI(auth *TwitterAuth) *anaconda.TwitterApi {
	// calc hash
	str := fmt.Sprintf("%v", *auth)
	s := md5.New()
	hash := s.Sum([]byte(str))

	if twitterAPI == nil || reflect.DeepEqual(apiHash, hash) == false {
		// (re)Authnication
		log.Println("TwitterAPI Authnication")
		log.Println("new authtoken hash : ", hash)
		anaconda.SetConsumerKey(auth.ConsumerKey)
		anaconda.SetConsumerSecret(auth.ConsumerSecret)
		twitterAPI = anaconda.NewTwitterApi(auth.AccessToken, auth.AccessSecret)
		apiHash = hash
	}
	return twitterAPI
}

func Tweet(message string, auth *TwitterAuth) (err error) {

	api := getTwitterAPI(auth)
	if api == nil {
		log.Println("Can't Get TwitterAPI Object")
		return *new(error)
	}
	tweet, err := api.PostTweet(message, nil)
	if err != nil {
		log.Println("Error while post tweet : ", err)
		return err
	}
	log.Println("tweet success")
	log.Println(tweet.Text)
	return nil
}

// Schedule
type Schedule struct {
	Event    eventdata.EventData
	Time     time.Time
	Message  string
	Executed bool
	Hash     []byte
}
type Schedules []Schedule

var hasher = md5.New()

func (s *Schedules) Append(e eventdata.EventData, t time.Time, msg string) {
	h := hasher.Sum([]byte(fmt.Sprintf("%v%v", e, t)))
	if !s.already(h) {
		*s = append(*s,
			Schedule{
				Event:    e,
				Time:     t,
				Message:  msg,
				Executed: false,
				Hash:     h,
			})
		log.Printf("Schedule append : %v\n%v\n", t.In(jst), msg)
	} else {
		log.Printf("Schedule append skip : %v\n%v\n", t.In(jst), msg)
	}
}

func (s *Schedules) already(hash []byte) bool {
	for _, t := range *s {
		if reflect.DeepEqual(t.Hash, hash) {
			return true
		}
		continue
	}
	return false
}
