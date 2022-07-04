package main

import (
	"bufio"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/tkanos/gonfig"
	"github.com/yannismate/yannismate-api/libs/rest/trackernet"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

var configuration = Configuration{}
var cliPlaylists = map[string]trackernet.Playlist{"unranked": trackernet.Unranked, "1v1": trackernet.Ranked1v1,
	"2v2": trackernet.Ranked2v2, "3v3": trackernet.Ranked3v3, "hoops": trackernet.Hoops,
	"rumble": trackernet.Rumble, "dropshot": trackernet.Dropshot, "snowday": trackernet.Snowday, "tournaments": trackernet.Tournaments}

func main() {
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		log.WithField("event", "load_config").Fatal(err)
		return
	}

	playlistStr := flag.String("playlist", "3v3", "Playlist (unranked, 1v1, 2v2, 3v3, hoops, rumble, dropshot, snowday, tournaments)")
	seasonFrom := flag.Int("fromseason", 20, "Fetch data starting from season (inclusive)")
	seasonTo := flag.Int("toseason", 21, "Fetch data until season (inclusive)")
	fileName := flag.String("in", "", "-in [in.csv]")

	flag.Parse()

	playlist, ok := cliPlaylists[*playlistStr]
	if !ok {
		log.Fatalf("Unknown playlist %v", playlist)
		return
	}

	if *seasonTo < *seasonFrom {
		log.Fatal("toseason cannot be smaller than fromseason")
		return
	}
	if *seasonTo < 0 || *seasonFrom < 0 {
		log.Fatal("Seasons cannot be negative")
		return
	}

	if *fileName == "" {
		log.Fatal("Usage: seeding -in [file.csv]")
		return
	}

	inFile, err := os.Open(*fileName)
	if err != nil {
		log.Fatal("Could not open file!", err)
		return
	}
	defer inFile.Close()

	outFile, err := os.Create("out.csv")
	if err != nil {
		log.Fatal("Could not create output file!", err)
		return
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	amountCols := 0
	lineNum := 0
	for scanner.Scan() {
		cols := len(strings.Split(scanner.Text(), ","))
		if cols%2 == 0 {
			log.Warnf("Unexpected amount of columns in line `%v` (%v)", lineNum, cols)
		}
		if amountCols < cols {
			amountCols = cols
		}
		lineNum++
	}

	_, _ = inFile.Seek(0, io.SeekStart)
	scanner = bufio.NewScanner(inFile)
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if firstLine {
			firstLine = false
			_, _ = outFile.WriteString(line + "\n")
			continue
		}

		splitted := strings.Split(line, ",")
		if len(splitted)%2 == 0 {
			log.Warnf("Unexpected amount of columns in line `%v`", line)
		}
		for i := len(splitted) - 1; i < amountCols; i++ {
			line = line + ","
		}
		for i := 2; i < len(splitted); i += 2 {
			playerName := splitted[i]
			if playerName == "" {
				line = line + strings.Repeat("No Epic ID,", (*seasonFrom+1)-*seasonTo)
			} else {
				mmrs := getMMRs(splitted[i], *seasonFrom, *seasonTo, playlist)
				line = line + strings.Join(mmrs, ",") + ","
			}
		}

		_, _ = outFile.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("File scanner error! ", err)
	}
}

func getMMRs(user string, seasonFrom int, seasonTo int, playlist trackernet.Playlist) []string {
	var results []string
	for i := seasonFrom; i <= seasonTo; i++ {
		log.Printf("Getting ranks for player %v in season %v", user, i)
		rankRes, err := GetRanks(user, i)
		if err != nil {
			if _, ok := err.(*TggError); ok {
				log.Warnf("Player with Epic ID %v was not found!", user)
				results = append(results, "No data found")
			} else {
				log.Errorf("Unknown Error fetching data for player %v: %v, try again after 20 seconds pause", user, err)
				time.Sleep(20 * time.Second)
				return getMMRs(user, seasonFrom, seasonTo, playlist)
			}
		}
		didAppend := false
		if err == nil {
			for _, rating := range rankRes.Rankings {
				if strings.Compare(string(rating.Playlist), string(playlist)) == 0 {
					results = append(results, strconv.Itoa(rating.Mmr))
					didAppend = true
					break
				}
			}
			if !didAppend {
				results = append(results, "No MMR")
			}
		}
	}

	return results
}
