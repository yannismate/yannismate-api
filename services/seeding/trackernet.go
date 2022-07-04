package main

import (
	"encoding/json"
	"github.com/yannismate/yannismate-api/libs/rest/trackernet"
	"github.com/yannismate/yannismate-api/libs/rest/webscraper"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var httpClient = http.Client{
	Timeout: time.Second * 20,
}

func GetRanks(user string, season int) (*trackernet.GetRankResponse, error) {

	requestUrl := strings.Replace(configuration.TrackerNet.BaseUrl, "$(name)", strings.Replace(url.QueryEscape(user), "+", "%20", -1), -1) + "?season=" + strconv.Itoa(season)
	req, err := http.NewRequest("GET", configuration.ScraperUrl+"?url="+url.QueryEscape(requestUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "yannismate-api/services/trackernet")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	scraperRes := webscraper.GetScrapeResponse{}
	err = json.Unmarshal(body, &scraperRes)
	if err != nil {
		return nil, err
	}

	tggRes := TggResponse{}

	err = json.Unmarshal([]byte(scraperRes.Content), &tggRes)
	if err != nil {
		return nil, err
	}

	if len(tggRes.Errors) > 0 {
		return nil, &TggError{}
	}

	rankings := make([]trackernet.Ranking, 0)
	for _, s := range tggRes.Data {
		if s.Type == "playlist" {
			ranking := s.toRanking()
			if ranking != nil {
				rankings = append(rankings, *ranking)
			}
		}
	}

	return &trackernet.GetRankResponse{Rankings: rankings}, nil
}

type TggResponse struct {
	Errors []map[string]interface{} `json:"errors"`
	Data   []TggSegment             `json:"data"`
}

type TggSegment struct {
	Type     string          `json:"type"`
	Metadata TggSegmentMeta  `json:"metadata"`
	Stats    TggSegmentStats `json:"stats"`
}

type TggSegmentMeta struct {
	Name string `json:"name"`
}

type TggSegmentStats struct {
	Tier     TggStatsValue `json:"tier"`
	Division TggStatsValue `json:"division"`
	Rating   TggStatsValue `json:"rating"`
}

type TggStatsValue struct {
	Value int `json:"value"`
}

type TggError struct{}

func (t TggError) Error() string {
	return "tracker.gg API returned error object"
}

var playlists = map[string]trackernet.Playlist{"Un-Ranked": trackernet.Unranked, "Ranked Duel 1v1": trackernet.Ranked1v1,
	"Ranked Doubles 2v2": trackernet.Ranked2v2, "Ranked Standard 3v3": trackernet.Ranked3v3, "Hoops": trackernet.Hoops,
	"Rumble": trackernet.Rumble, "Dropshot": trackernet.Dropshot, "Snowday": trackernet.Snowday, "Tournament Matches": trackernet.Tournaments}

func (seg *TggSegment) toRanking() *trackernet.Ranking {

	playlist, ok := playlists[seg.Metadata.Name]
	if !ok {
		return nil
	}

	return &trackernet.Ranking{
		Playlist: playlist,
		Mmr:      seg.Stats.Rating.Value,
		Rank:     seg.Stats.Tier.Value,
		Division: seg.Stats.Division.Value,
	}
}
