package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yannismate/yannismate-api/libs/rest/trackernet"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func GetRankString(platform string, user string, format string) (string, error) {

	res, err := requestRank(platform, user)
	if err != nil {
		return "", err
	}

	return formatRankResponse(res, format), nil
}

type PlayerNotFoundError struct{}

func (p PlayerNotFoundError) Error() string {
	return "Player not found"
}

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

func requestRank(platform string, user string) (*trackernet.GetRankResponse, error) {

	reqUrl := configuration.TrackerNetServiceUrl + "/rank?platform=" + platform + "&user=" + url.QueryEscape(user)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		log.WithField("event", "new_request_trackernet").Error(err)
		return nil, err
	}
	req.Header.Set("User-Agent", "yannismate-api/services/twitchbot")

	res, err := httpClient.Do(req)
	if err != nil {
		log.WithField("event", "do_request_trackernet").Error(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, &PlayerNotFoundError{}
	}

	var rankRes trackernet.GetRankResponse
	err = json.NewDecoder(res.Body).Decode(&rankRes)
	if err != nil {
		log.WithField("event", "read_body_trackernet").Error(err)
		return nil, err
	}

	return &rankRes, nil
}

var tokenMatcher = regexp.MustCompile("\\$\\(((\\w|\\.)+)\\)")

func formatRankResponse(response *trackernet.GetRankResponse, format string) string {

	var result strings.Builder

	matchesBytes := tokenMatcher.FindAllStringIndex(format, -1)

	var matches [][]int
	for _, x := range matchesBytes {
		matches = append(matches, []int{utf8.RuneCountInString(format[:x[0]]), utf8.RuneCountInString(format[:x[1]])})
	}

	if len(matches) == 0 {
		return format
	}

	formatChars := []rune(format)
	nextMatch := matches[0]
	matchIndex := 0

	for i, ch := range formatChars {
		if i == nextMatch[0] {
			// insert token for current match
			token := formatChars[nextMatch[0]+2 : nextMatch[1]-1]
			result.WriteString(evalToken(response, string(token)))
		} else if i+1 == nextMatch[1] {
			// use next match
			matchIndex++
			if matchIndex >= len(matches) {
				nextMatch = []int{-1, -1}
			} else {
				nextMatch = matches[matchIndex]
			}
		} else if i > nextMatch[0] && i < nextMatch[1] {
			// skip token chars
			continue
		} else {
			// non token char
			result.WriteRune(ch)
		}
	}

	return result.String()
}

var tokenExtractor = regexp.MustCompile("^([u123hrdst])\\.([rdm])\\.?([sml])?$")

func evalToken(response *trackernet.GetRankResponse, token string) string {
	if token == "name" {
		return response.DisplayName
	}

	matches := tokenExtractor.FindAllStringSubmatch(token, -1)
	if len(matches) == 0 {
		return "${" + string(token) + "}"
	}

	playlist := playlistFromAbbr(matches[0][1])
	stat := matches[0][2]
	modifier := matches[0][3]

	if modifier == "" {
		modifier = "l"
	}

	for _, ranking := range response.Rankings {
		if playlist == ranking.Playlist {
			if stat == "r" {
				return rankToStr(ranking.Rank, modifier)
			} else if stat == "d" {
				if modifier == "l" || modifier == "m" {
					return toRoman(ranking.Division + 1)
				} else {
					return strconv.Itoa(ranking.Division + 1)
				}
			} else if stat == "m" {
				return strconv.Itoa(ranking.Mmr)
			}
		}
	}

	return "[err:" + token + "]"
}

func playlistFromAbbr(abbr string) trackernet.Playlist {
	switch abbr {
	case "u":
		return trackernet.Unranked
	case "1":
		return trackernet.Ranked1v1
	case "2":
		return trackernet.Ranked2v2
	case "3":
		return trackernet.Ranked3v3
	case "h":
		return trackernet.Hoops
	case "r":
		return trackernet.Rumble
	case "d":
		return trackernet.Dropshot
	case "s":
		return trackernet.Snowday
	case "t":
		return trackernet.Tournaments
	}
	return trackernet.Unranked
}

var ranksS = map[int]string{
	0: "UR", 1: "B1", 2: "B2", 3: "B3", 4: "S1", 5: "S2", 6: "S3", 7: "G1", 8: "G2", 9: "G3",
	10: "P1", 11: "P2", 12: "P3", 13: "D1", 14: "D2", 15: "D3",
	16: "C1", 17: "C2", 18: "C3", 19: "GC1", 20: "GC2",
	21: "GC3", 22: "SSL",
}
var ranksM = map[int]string{
	0: "Unranked", 1: "Bronze I", 2: "Bronze II", 3: "Bronze III",
	4: "Silver I", 5: "Silver II", 6: "Silver III", 7: "Gold I", 8: "Gold II", 9: "Gold III",
	10: "Plat I", 11: "Plat II", 12: "Plat III", 13: "Dia I", 14: "Dia II", 15: "Dia III",
	16: "Champ I", 17: "Champ II", 18: "Champ III", 19: "Grand Champ I", 20: "Grand Champ II",
	21: "Grand Champ III", 22: "SSL",
}
var ranksL = map[int]string{
	0: "Unranked", 1: "Bronze I", 2: "Bronze II", 3: "Bronze III",
	4: "Silver I", 5: "Silver II", 6: "Silver III", 7: "Gold I", 8: "Gold II", 9: "Gold III",
	10: "Platinum I", 11: "Platinum II", 12: "Platinum III", 13: "Diamond I", 14: "Diamond II", 15: "Diamond III",
	16: "Champion I", 17: "Champion II", 18: "Champion III", 19: "Grand Champion I", 20: "Grand Champion II",
	21: "Grand Champion III", 22: "Supersonic Legend",
}

func rankToStr(rank int, modifier string) string {
	if rank > 22 {
		return "?"
	}
	if modifier == "s" {
		return ranksS[rank]
	} else if modifier == "m" {
		return ranksM[rank]
	} else if modifier == "l" {
		return ranksL[rank]
	}
	return "??"
}

func toRoman(num int) string {
	switch num {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	case 4:
		return "IV"
	}
	return "?"
}
