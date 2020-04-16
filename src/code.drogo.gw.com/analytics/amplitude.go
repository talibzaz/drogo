package analytics

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
	"errors"
	"strings"
	"strconv"
)

const (
	methodGet = "GET"
)

type analyticsClient struct {
	apiKey 	string
	secretKey 	string
}

func newAnalyticsAuthClient() *analyticsClient{
	return &analyticsClient{
		apiKey: "def6005810e873f9bb90521922ad50d1",
		secretKey: "759d6b05d8430bf4357654297bd75645",
	}
}

type clientAmplitudeId struct {
	Matches []struct {
		UserID      string `json:"user_id"`
		AmplitudeID int64  `json:"amplitude_id"`
	} `json:"matches"`
}

type eventDetails struct {
	UserData struct {
		NumEvents            int           `json:"num_events"`
	} `json:"userData"`
}

type amplitudeData struct {
	Data struct {
		SeriesLabels	[][]interface{}	`json:"seriesLabels"`
		SeriesCollapsed [][]struct{
			Value int	`json:"value"`
		} `json:"seriesCollapsed"`
	} `json:"data"`
}

func (c *analyticsClient) amplitudeRequest(req *http.Request) ([]byte, error){
	req.SetBasicAuth(c.apiKey, c.secretKey)

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if 200 != res.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}

	return body, nil
}

//PageViews returns the number of page hits of the given event / page.
func EventViews(duration, eventName string) (map[string]int, error){
	var startDate time.Time
	endDate := time.Now()

	switch strings.ToLower(duration) {
	case "week":
		startDate = endDate.AddDate(0,0,-7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		return nil, errors.New("invalid date duration")
	}

	//Convert the start and end date into string of the layout => 20180101
	startTime := strings.SplitAfter(startDate.String(), " ")
	start := strings.Replace(startTime[0],"-", "", -1)
	start = strings.TrimSpace(start)

	endTime := strings.SplitAfter(endDate.String(), " ")
	end := strings.Replace(endTime[0], "-", "", -1)
	end = strings.TrimSpace(end)

	eventName = strings.TrimSpace(eventName)
	url := `https://amplitude.com/api/2/events/segmentation?e={"event_type":"`+eventName+`","group_by":[{"type":"user","value":"user_id"}]}&m=totals&start=`+start+`&end=`+end+`&limit=5&i=1`


	req, err := http.NewRequest(methodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := newAnalyticsAuthClient()

	body, err := client.amplitudeRequest(req)
	if err != nil {
		return nil, err
	}

	var data amplitudeData

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	eventHits := make(map[string]int, 0)

	for i, v := range data.Data.SeriesLabels{
		eventHits[v[1].(string)] = data.Data.SeriesCollapsed[i][0].Value
	}

	return eventHits, nil
}

//Page Views for single event
func GetPageViewsByEventID(eventId string) (int, error) {
	client := newAnalyticsAuthClient()

	res, err := getAmplitudeId(eventId,client)
	if err != nil {
		return 0, err
	}

	amplitudeID := strconv.FormatInt(res, 10)

	eventHits, err := getEventPageHits(amplitudeID, client)
	if err != nil {
		return 0, err
	}

	return eventHits, nil
}

func getAmplitudeId(eventId string, client *analyticsClient) (int64, error){
	URL := `https://amplitude.com/api/2/usersearch?user=` + eventId

	req, err := http.NewRequest(methodGet, URL, nil)
	if err != nil {
		return 0, err
	}

	body, err := client.amplitudeRequest(req)
	if err != nil {
		return 0, err
	}

	var clientData clientAmplitudeId

	err = json.Unmarshal(body, &clientData)
	if err != nil {
		return 0, err
	}

	return clientData.Matches[0].AmplitudeID, nil
}

func getEventPageHits(amplitudeID string, client *analyticsClient) (int,error) {
	URL := `https://amplitude.com/api/2/useractivity?user=` + amplitudeID

	req, err := http.NewRequest(methodGet, URL, nil)
	if err != nil {
		return 0, err
	}

	body, err := client.amplitudeRequest(req)
	if err != nil {
		return 0, err
	}

	var eventHits eventDetails

	err = json.Unmarshal(body, &eventHits)
	if err != nil {
		return 0, err
	}

	return eventHits.UserData.NumEvents, nil
}