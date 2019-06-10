package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	. "time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

func getEventsResponse(yearMonth Time, limit int) []byte {
	baseUrl := "http://eventon.jp/api/events.json"
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	query := u.Query()
	query.Add("ym", fmt.Sprintf("%4d%02d", yearMonth.Year(), int(yearMonth.Month())))
	query.Add("limit", strconv.Itoa(limit))
	u.RawQuery = query.Encode()
	fmt.Println(u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bodyBytes
}

func parseResponse(response []byte) Events {
	var events Events
	err := json.Unmarshal(response, &events)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", events)

	return events
}

type Events struct {
	Events []Event `json:"events"`
}

type Event struct {
	EventId   string `json:"event_id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	StartedAt Time   `json:"started_at"`
	EventUrl  string `json:"event_url"`
	Address   string `json:"address"`
	ImagePath string `json:"image_path"`
}

func yearMonth(year, month int) Time {
	return Date(year, Month(month), 1, 0, 0, 0, 0, UTC)
}

func addMonth(t Time, month int) Time {
	return t.AddDate(0, month, 0)
}

func main() {
	var events []Event
	for t := yearMonth(2018, 1); t.Before(yearMonth(2020, 1)); t = addMonth(t, 1) {
		eventsByte := getEventsResponse(t, 100)
		e := parseResponse(eventsByte).Events
		events = append(events, e...)
	}

	db := dynamo.New(session.New(), &aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	table := db.Table("event2")

	for _, event := range events {
		fmt.Println("adding event:", event)
		if err := table.Put(event).Run(); err != nil {
			fmt.Println(err)
		}
	}
}
