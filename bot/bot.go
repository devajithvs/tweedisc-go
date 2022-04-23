package bot

import (
	"strings"
	"tweedisc-go/config"
	"tweedisc-go/database"
	"tweedisc-go/embed"
	"tweedisc-go/ratelimiter"
	"tweedisc-go/twitter"

	"github.com/bwmarrin/discordgo"
)

var log = config.Log
var BotId string
var goBot *discordgo.Session

var validReactions = map[string]string{
	"❤️": "like0",
	"💙":  "like1",
	"🔁":  "retweet0",
	"🔄":  "retweet1",
}

var likeLimiter = ratelimiter.RateLimiter{
	Period:   config.Period,
	MaxCalls: config.LikeLimit,
}

var retweetLimiter = ratelimiter.RateLimiter{
	Period:   config.Period,
	MaxCalls: config.RetweetLimit,
}

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	BotId = u.ID

	goBot.AddHandler(messageHandler)
	goBot.AddHandler(reactionAddHandler)
	goBot.AddHandler(reactionRemoveHandler)

	err = goBot.Open()

	if err != nil {
		log.Fatal(err.Error())
		return
	}
	log.Println("Bot is running !")
	addReactionToExistingMessages(goBot)

}

func addReactionToExistingMessages(s *discordgo.Session) {
	for _, guild := range s.State.Guilds {
		channels, _ := s.GuildChannels(guild.ID)

		for _, c := range channels {
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			messages, err := s.ChannelMessages(c.ID, 20, "", "", "")
			if err != nil {
				continue
			}

			for _, m := range messages {

				if twitter.IsTweet(m.Content) && len(m.Reactions) == 0 {
					messageReaction(s, m)
				} else {
					break
				}
			}

		}
	}
}

func sendTweetMessage(s *discordgo.Session, userID string, title string, tweetLink string) error {
	embedMessage := embed.NewEmbed().SetTitle(title).SetDescription(tweetLink)
	author := twitter.GetTweetAuthor(tweetLink)
	embedMessage.SetAuthor(author, tweetLink)
	embedMessage.SetColor(8808703)
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}
	s.ChannelMessageSendEmbed(channel.ID, embedMessage.MessageEmbed)
	return nil
}

func sendAuthLink(s *discordgo.Session, userID string, guildID string) error {
	server, err := database.GetServer(guildID)
	var loginURL string
	if err != nil {
		loginURL = config.GumroadURI
	} else {
		loginURL = server.Gumroad_url
	}
	log.Println(loginURL, err)
	embedMessage := embed.NewEmbed().SetTitle("⛔ You have not enabled Tweedisc feature").SetDescription("Please click __**[HERE](" + loginURL + ")**__ to enable Tweedisc feature (engaging with tweets directly from Discord).")
	embedMessage.SetColor(8808703)
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}
	s.ChannelMessageSendEmbed(channel.ID, embedMessage.MessageEmbed)
	return nil
}

func sendErrorMessage(s *discordgo.Session, userID string, title string) error {
	embedMessage := embed.NewEmbed().SetTitle("⛔ " + title)
	embedMessage.SetColor(8808703)

	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}
	s.ChannelMessageSendEmbed(channel.ID, embedMessage.MessageEmbed)
	return nil
}

func reactionAddHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == BotId {
		return
	}

	m, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	log.Println("Add reaction: ", m.Author.Username)

	if err != nil {
		s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
		log.Error("Error getting message: ", err)
		return
	}

	action, exist := validReactions[r.Emoji.Name]
	if !exist {
		log.Warn("Action doesn't exist: ", r.Emoji.Name)
		return
	}

	if twitter.IsTweet(m.Content) {
		user, err := database.GetUser(r.UserID)
		if err != nil || user.Created_time_stamp == 0 {
			log.Warn(err)
			s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
			// sendAuthLink(s, r.UserID, r.GuildID)
			return
		}
		tweetLinks := twitter.GetTweetLinks(m.Content)
		index := int(action[len(action)-1] - '0')
		if index >= len(tweetLinks) {
			s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
			// sendErrorMessage(s, r.UserID, "Invalid Tweet")
			return
		}
		if strings.HasPrefix(action, "like") {
			if likeLimiter.CheckLimit(user.Twitter_user_id, "like") {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
				// sendErrorMessage(s, r.UserID, "You have reached the maximum number of requests for likes per 15 minutes. Please try after some time.")
				return
			}
			err := twitter.LikeTweet(tweetLinks[index], user)
			if err != nil {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
				// sendErrorMessage(s, r.UserID, "Error. Please try after some time")
				return
			}
			sendTweetMessage(s, r.UserID, "✅ Successfully liked", tweetLinks[index])

		} else if strings.HasPrefix(action, "retweet") {
			if retweetLimiter.CheckLimit(user.Twitter_user_id, "retweet") {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
				// sendErrorMessage(s, r.UserID, "You have reached the maximum number of requests for retweets per 15 minutes. Please try after some time.")
				return
			}
			err := twitter.RetweetTweet(tweetLinks[index], user)
			if err != nil {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
				// sendErrorMessage(s, r.UserID, "Error. Please try after some time")
				return
			}
			sendTweetMessage(s, r.UserID, "✅ Successfully retweeted", tweetLinks[index])

		}
	}
}

func reactionRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.UserID == BotId {
		return
	}

	m, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	log.Println("Remove reaction: ", m.Author.Username)

	if err != nil {
		log.Error("Error getting message: ", err)
		return
	}

	action, exist := validReactions[r.Emoji.Name]
	if !exist {
		log.Warn("Action doesn't exist: ", r.Emoji.Name)
		return
	}

	if twitter.IsTweet(m.Content) {
		user, err := database.GetUser(r.UserID)
		if err != nil || user.Created_time_stamp == 0 {
			log.Warn(err)
			sendAuthLink(s, r.UserID, r.GuildID)
			return
		}
		tweetLinks := twitter.GetTweetLinks(m.Content)
		index := int(action[len(action)-1] - '0')
		if index >= len(tweetLinks) {
			return
		}
		if strings.HasPrefix(action, "like") {
			if likeLimiter.CheckLimit(user.Twitter_user_id, "like") {
				sendErrorMessage(s, r.UserID, "You have reached the maximum number of requests for likes per 15 minutes. Please try after some time.")
				return
			}
			err := twitter.UnlikeTweet(tweetLinks[index], user)
			if err != nil {
				sendErrorMessage(s, r.UserID, "Error. Please try after some time")
				return
			}
			sendTweetMessage(s, r.UserID, "✅ Successfully unliked", tweetLinks[index])

		} else if strings.HasPrefix(action, "retweet") {
			if retweetLimiter.CheckLimit(user.Twitter_user_id, "retweet") {
				sendErrorMessage(s, r.UserID, "You have reached the maximum number of requests for retweets per 15 minutes. Please try after some time.")
				return
			}
			err := twitter.UnRetweetTweet(tweetLinks[index], user)
			if err != nil {
				sendErrorMessage(s, r.UserID, "Error. Please try after some time")
				return
			}
			sendTweetMessage(s, r.UserID, "✅ Successfully unretweeted", tweetLinks[index])
		}
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotId {
		return
	}
	tweetLinks := twitter.GetTweetLinks(m.Content)
	if len(tweetLinks) == 1 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❤️")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔁")
	} else if len(tweetLinks) > 1 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❤️")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔁")
		s.MessageReactionAdd(m.ChannelID, m.ID, "💙")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔄")
	}

	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func messageReaction(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID == BotId {
		return
	}
	tweetLinks := twitter.GetTweetLinks(m.Content)
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return
	}
	log.Println("Adding reactions to " + m.ID + " in " + guild.Name + " in channel " + channel.Name)
	log.Println(m.Content)
	if len(tweetLinks) == 1 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❤️")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔁")
	} else if len(tweetLinks) > 1 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "❤️")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔁")
		s.MessageReactionAdd(m.ChannelID, m.ID, "💙")
		s.MessageReactionAdd(m.ChannelID, m.ID, "🔄")
	}
}
