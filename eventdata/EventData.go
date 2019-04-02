package eventdata

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"time"
)

type rawEventDatas struct {
	EventData []rawEventData `json:"events"`
}

type rawEventData struct {
	ID           int         `json:"id"`
	URL          string      `json:"url"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	StartDateStr string      `json:"utc_start_date"`
	EndDateStr   string      `json:"utc_end_date"`
	Organizer    []organizer `json:"organizer"`
}

type organizer struct {
	ID        int    `json:"id"`
	Organizer string `json:"organizer"`
	Slug      string `json:"slug"`
}

type EventData struct {
	ID          int
	URL         string
	Title       string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	Organizer   string
}

var jst, _ = time.LoadLocation("Asia/Tokyo")

func GetEventsFromWordpress(url string, dayLineHour int) (events []EventData, err error) {
	nowt := time.Now()
	nowt = time.Date(nowt.Year(), nowt.Month(), nowt.Day(), dayLineHour, 0, 0, 0, jst)
	nowt = nowt.Add(time.Hour * 24 * 3)
	sdt := nowt.Format("2006/01/02T15:04")
	edt := nowt.Add(2 * 24 * time.Hour).Add(1 * time.Minute).Format("2006/01/02T15:04")
	url = "https://" + url + "/?rest_route=/tribe/events/v1/events" + "&start_date=" + sdt + "&end_date=" + edt
	log.Println("Getting URL : ", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("HTTP Get error:", err)
		return
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("HTTP Read error:", err)
		return
	}

	return GetEventDatas(byteArray)
}

var eventsHash []byte
var datasCache []EventData

func GetEventDatas(jsonBytes []byte) (datas []EventData, err error) {
	s := md5.New()
	hash := s.Sum(jsonBytes)
	if reflect.DeepEqual(eventsHash, hash) {
		log.Println("GetEventDatas return cache")
		return datasCache, nil
	}

	rawdatas := new(rawEventDatas)
	err = json.Unmarshal(jsonBytes, rawdatas)
	if err != nil {
		log.Println("JSON Unmarshal error:", err)
		return
	}

	for _, rawdata := range rawdatas.EventData {
		d, err := parseEventData(&rawdata)
		if err != nil {
			log.Println("EventData Parse Error:", err)
			continue
		}
		datas = append(datas, d)
	}
	log.Println("Update Events : ", datas)
	eventsHash = hash
	datasCache = datas
	return datasCache, nil
}

func GetEventData(jsonBytes []byte) (data EventData, err error) {
	rawdata := new(rawEventData)
	err = json.Unmarshal(jsonBytes, rawdata)
	if err != nil {
		log.Println("JSON Unmarshal error:", err)
		return
	}

	return parseEventData(rawdata)
}

func parseEventData(rawdata *rawEventData) (data EventData, err error) {
	data.ID = rawdata.ID
	parsedurl, err := url.Parse(rawdata.URL)
	if err != nil {
		data.URL = rawdata.URL
	} else {
		data.URL = fmt.Sprintf("%s://%s?p=%d", parsedurl.Scheme, parsedurl.Host, data.ID)
	}

	// TODO : trim for post
	data.Title = rawdata.Title
	data.Description = rawdata.Description

	data.Description = parseDescription(rawdata.Description)

	data.StartDate, _ = time.Parse("2006-01-02 15:04:05", rawdata.StartDateStr)
	data.EndDate, _ = time.Parse("2006-01-02 15:04:05", rawdata.EndDateStr)
	if len(rawdata.Organizer) == 0 {
		data.Organizer = ""
	} else {
		data.Organizer = rawdata.Organizer[0].Organizer
	}
	return
}

func parseDescription(raw string) (parsed string) {
	r := regexp.MustCompile(`<(.+?)>(.+?)<.+?>`)
	br := regexp.MustCompile(`(p)|(br.*)|(h[0-9])`)
	b := regexp.MustCompile(`^(h¥d)|(strong)|(b)¥s*$`)
	tmp := r.FindAllStringSubmatch(raw, -1)
	for _, t := range tmp {
		// add prefixes
		if b.MatchString(t[1]) {
			t[2] = "*" + t[2] + "*"
		}

		// print string
		if br.MatchString(t[1]) {
			parsed += fmt.Sprintln(t[2])
		} else {
			parsed += fmt.Sprint(t[2])
		}
	}
	log.Print(parsed)

	return parsed
}
