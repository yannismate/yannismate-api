package main

import (
	"encoding/json"
	"github.com/gempir/go-twitch-irc/v3"
	log "github.com/sirupsen/logrus"
	"github.com/tkanos/gonfig"
	"github.com/yannismate/yannismate-api/libs/cache"
	"github.com/yannismate/yannismate-api/libs/rest/trackernet"
	"strconv"
	"strings"
	"time"
)

var configuration Configuration
var botDb *BotDb
var redisCache cache.Cache

func main() {
	log.Info("Starting twitchbot...")
	err := gonfig.GetConf("config.json", &configuration)
	if err != nil {
		log.WithField("event", "load_config").Fatal(err)
		return
	}

	botDb, err = NewBotDb(configuration.DbUri)
	if err != nil {
		log.WithField("event", "connect_db").Fatal(err)
		return
	}

	client := twitch.NewClient(configuration.TwitchUsername, configuration.TwitchToken)
	client.SetJoinRateLimiter(twitch.CreateDefaultRateLimiter())

	redisCache = cache.NewCache(configuration.CacheUrl)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		go handleMessage(message, client)
	})
	client.Join(configuration.TwitchUsername)

	client.OnConnect(func() {
		log.WithField("event", "irc_connected").Info("IRC connected")
		namesCursor := ""
		for {
			names, newNamesCursor, err := botDb.GetUserNames(namesCursor, 20)
			if newNamesCursor != nil {
				namesCursor = *newNamesCursor
			}
			if err != nil {
				log.WithField("event", "channels_rejoin").Fatal(err)
				return
			}

			if len(names) > 0 {
				log.WithField("event", "channels_rejoin").Info("Joining " + strconv.Itoa(len(names)) + " channels")
				client.Join(names...)
			}

			if newNamesCursor == nil || len(names) < 20 {
				break
			}
			time.Sleep(time.Second * 10)
		}
		client.Say(configuration.TwitchUsername, configuration.TwitchUsername+" online! MrDestructoid")
	})

	err = client.Connect()
}

func handleMessage(message twitch.PrivateMessage, client *twitch.Client) {
	if message.Channel == strings.ToLower(configuration.TwitchUsername) {
		switch strings.Split(strings.ToLower(message.Message), " ")[0] {
		case "!join":
			joinChannelCommand(&message, client)
		case "!leave":
			leaveChannelCommand(&message, client)
		case "!setplatform":
			setPlatformCommand(&message, client)
		case "!setusername":
			setUsernameCommand(&message, client)
		case "!setformat":
			setFormatCommand(&message, client)
		}
	} else if strings.HasPrefix(strings.ToLower(message.Message), "@"+strings.ToLower(configuration.TwitchUsername)) {
		if message.Channel == message.User.Name {
			msg := strings.TrimPrefix(strings.ToLower(message.Message), "@"+strings.ToLower(configuration.TwitchUsername)+" ")
			switch strings.Split(msg, " ")[0] {
			case "!leave":
				leaveChannelCommand(&message, client)
			case "!setplatform":
				setPlatformCommand(&message, client)
			case "!setusername":
				setUsernameCommand(&message, client)
			case "!setformat":
				setFormatCommand(&message, client)
			}
		} else {
			isMod, ok := message.Tags["mod"]
			if ok && isMod == "1" {
				msg := strings.TrimPrefix(strings.ToLower(message.Message), "@"+strings.ToLower(configuration.TwitchUsername)+" ")
				switch strings.Split(msg, " ")[0] {
				case "!setplatform":
					setPlatformCommand(&message, client)
				case "!setusername":
					setUsernameCommand(&message, client)
				case "!setformat":
					setFormatCommand(&message, client)
				}
			}
		}
	} else if strings.HasPrefix(message.Message, "!") {
		checkForRankCommand(&message, client)
	}
}

func joinChannelCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	log.WithField("event", "join_command").WithField("channel", message.Channel).Info("Executing join command")
	_, err := botDb.GetBotUserByTwitchUserId(message.User.ID)
	if err == nil {
		client.Say(message.Channel, "@"+message.User.Name+" The bot has already joined your channel.")
		return
	}
	newUser := BotUser{
		TwitchUserId:      message.User.ID,
		TwitchLogin:       message.User.Name,
		TwitchCommandName: "rank",
		RlMessageFormat:   "Ranked 1v1: $(1.r) Div $(1.d) ($(1.m)) | Ranked 2v2: $(2.r) Div $(2.d) ($(2.m)) | Ranked 3v3: $(3.r) Div $(3.d) ($(3.m)) | Tournament Matches: $(t.r) Div $(t.d) ($(t.m))",
	}
	err = botDb.InsertBotUser(newUser)
	if err != nil {
		client.Say(message.Channel, "@"+message.User.Name+" Error joining channel "+message.User.Name)
		log.WithField("event", "join_command").Error(err)
		return
	}
	client.Join(message.User.Name)
	client.Say(message.Channel, "@"+message.User.Name+" The bot has now joined your channel!")
}

func leaveChannelCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	log.WithField("event", "leave_command").WithField("channel", message.Channel).Info("Executing leave command")
	wasDeleted, err := botDb.DeleteBotUserByTwitchUserId(message.User.ID)
	if err != nil {
		client.Say(message.Channel, "@"+message.User.Name+" Error leaving channel "+message.User.Name)
		log.WithField("event", "leave_command").Error(err)
		return
	}
	if wasDeleted {
		redisCache.Delete("twitch:" + message.Channel)
		client.Say(message.Channel, "@"+message.User.Name+" Leaving channel "+message.User.Name)
		client.Depart(message.User.Name)
	} else {
		client.Say(message.Channel, "@"+message.User.Name+" The bot was not joined to channel "+message.User.Name)
	}
}

var validPlatforms = map[string]bool{
	trackernet.Epic:  true,
	trackernet.Steam: true,
	trackernet.PS:    true,
	trackernet.Xbox:  true,
}

func setPlatformCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	log.WithField("event", "setplatform_command").WithField("channel", message.Channel).Info("Executing setplatform command")
	cmdContent := strings.SplitN(message.Message, "!setplatform ", 2)
	if len(cmdContent) != 2 {
		return
	}
	newPlatform := cmdContent[1]

	if !validPlatforms[newPlatform] {
		client.Say(message.Channel, "@"+message.User.Name+" Valid platforms: epic, steam, ps, xbox")
		return
	}

	var user string
	if message.Channel == configuration.TwitchUsername {
		user = message.User.Name
	} else {
		user = message.Channel
	}

	wasChanged, err := botDb.UpdateRlPlatformByTwitchLogin(user, newPlatform)
	if err != nil {
		client.Say(message.Channel, "@"+message.User.Name+" There was an error updating the platform")
		log.WithField("event", "setplatform_command_db_update").Error(err)
		return
	}
	if !wasChanged {
		client.Say(message.Channel, "@"+message.User.Name+" The bot is not joined")
		return
	}
	redisCache.Delete("twitch:" + message.Channel)
	client.Say(message.Channel, "@"+message.User.Name+" Platform updated")
}

func setUsernameCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	log.WithField("event", "setusername_command").WithField("channel", message.Channel).Info("Executing setusername command")
	cmdContent := strings.SplitN(message.Message, "!setusername ", 2)
	if len(cmdContent) != 2 {
		return
	}
	newUsername := cmdContent[1]

	var user string
	if message.Channel == configuration.TwitchUsername {
		user = message.User.Name
	} else {
		user = message.Channel
	}

	wasChanged, err := botDb.UpdateRlUsernameByTwitchLogin(user, newUsername)
	if err != nil {
		client.Say(message.Channel, "@"+message.User.Name+" There was an error updating the username")
		log.WithField("event", "setusername_command_db_update").Error(err)
		return
	}
	if !wasChanged {
		client.Say(message.Channel, "@"+message.User.Name+" The bot is not joined")
		return
	}
	redisCache.Delete("twitch:" + message.Channel)
	client.Say(message.Channel, "@"+message.User.Name+" Username updated")
}

func setFormatCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	log.WithField("event", "setformat_command").WithField("channel", message.Channel).Info("Executing setformat command")
	cmdContent := strings.SplitN(message.Message, "!setformat ", 2)
	if len(cmdContent) != 2 {
		return
	}
	newFormat := cmdContent[1]

	var user string
	if message.Channel == configuration.TwitchUsername {
		user = message.User.Name
	} else {
		user = message.Channel
	}

	wasChanged, err := botDb.UpdateRlMsgFormatByTwitchLogin(user, newFormat)
	if err != nil {
		client.Say(message.Channel, "@"+message.User.Name+" There was an error updating the format")
		log.WithField("event", "setformat_command_db_update").Error(err)
		return
	}
	if !wasChanged {
		client.Say(message.Channel, "@"+message.User.Name+" The bot is not joined")
		return
	}
	redisCache.Delete("twitch:" + message.Channel)
	client.Say(message.Channel, "@"+message.User.Name+" Format updated")
}

type CachedChannel struct {
	Command       string `json:"cmd"`
	LastExecuted  int64  `json:"last"`
	RlPlatform    string `json:"rlp"`
	RlUsername    string `json:"rlu"`
	MessageFormat string `json:"fmt"`
}

func checkForRankCommand(message *twitch.PrivateMessage, client *twitch.Client) {
	cachedStr, err := redisCache.Get("twitch:" + message.Channel)
	if err == nil {
		cachedObj := CachedChannel{}
		err = json.Unmarshal([]byte(cachedStr), &cachedObj)
		if err != nil {
			return
		}
		if cachedObj.LastExecuted > time.Now().Unix()-10 || !strings.HasPrefix(message.Message, "!"+cachedObj.Command) {
			return
		}
		log.WithField("event", "rank_command").WithField("channel", message.Channel).Info("Executing rank command")

		replyStr, err := GetRankString(cachedObj.RlPlatform, cachedObj.RlUsername, cachedObj.MessageFormat)
		if err != nil {
			if _, ok := err.(*PlayerNotFoundError); ok {
				client.Say(message.Channel, "Player "+cachedObj.RlUsername+" was not found on platform "+cachedObj.RlPlatform)
				return
			}
			client.Say(message.Channel, "There was an error getting the rank for player "+cachedObj.RlUsername+" on platform "+cachedObj.RlPlatform)
			log.WithField("event", "user_command_get_rank_str").Error(err)
		} else {
			client.Say(message.Channel, replyStr)
		}

		cachedObj.LastExecuted = time.Now().Unix()
		toCacheStr, _ := json.Marshal(cachedObj)
		err = redisCache.SetWithTtl("twitch:"+message.Channel, string(toCacheStr), time.Hour)
		if err != nil {
			log.WithField("event", "user_command_cache_set").Error(err)
		}
		return
	}

	dbUser, err := botDb.GetBotUserByTwitchLogin(message.Channel)
	if err != nil {
		log.WithField("event", "user_command_get_db").Warn(err)
		return
	}

	if message.Message == "!"+dbUser.TwitchCommandName {

		if dbUser.RlPlatform == "" {
			client.Say(message.Channel, "Please set your platform with \"@"+configuration.TwitchUsername+" !setplatform platform\"")
			return
		}
		if dbUser.RlUsername == "" {
			client.Say(message.Channel, "Please set your username with \"@"+configuration.TwitchUsername+" !setusername platform\"")
			return
		}

		log.WithField("event", "rank_command").WithField("channel", message.Channel).Info("Executing rank command")

		toCache := CachedChannel{
			Command:       dbUser.TwitchCommandName,
			LastExecuted:  time.Now().Unix(),
			RlPlatform:    dbUser.RlPlatform,
			RlUsername:    dbUser.RlUsername,
			MessageFormat: dbUser.RlMessageFormat,
		}
		toCacheStr, _ := json.Marshal(toCache)
		err := redisCache.SetWithTtl("twitch:"+message.Channel, string(toCacheStr), time.Hour)
		if err != nil {
			log.WithField("event", "user_command_cache_set").Error(err)
		}

		rankString, err := GetRankString(dbUser.RlPlatform, dbUser.RlUsername, dbUser.RlMessageFormat)
		if err != nil {
			if _, ok := err.(*PlayerNotFoundError); ok {
				client.Say(message.Channel, "Player "+dbUser.RlUsername+" was not found on platform "+dbUser.RlPlatform)
				return
			}
			client.Say(message.Channel, "There was an error getting the rank for player "+dbUser.RlUsername+" on platform "+dbUser.RlPlatform)
			log.WithField("event", "user_command_get_rank_str").Error(err)
			return
		}

		client.Say(message.Channel, rankString)
	}
}