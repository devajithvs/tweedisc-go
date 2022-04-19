package bot

import (
	"strings"
	"tweedisc-go/config"
	"tweedisc-go/embed"
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
}

func sendTweetMessage(s *discordgo.Session, userID string, title string, tweetLink string) error {
	embedMessage := embed.NewEmbed().SetTitle(title).SetDescription(tweetLink)
	author := twitter.GetTweetAuthor(tweetLink)
	embedMessage.SetAuthor(author, tweetLink)
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
		log.Error("Error getting message: ", err)
		return
	}

	action, exist := validReactions[r.Emoji.Name]
	if !exist {
		log.Warn("Action doesn't exist: ", r.Emoji.Name)
		return
	}
	if twitter.IsTweet(m.Content) {
		tweetLinks := twitter.GetTweetLinks(m.Content)
		index := action[len(action)-1] - '0'
		if strings.HasPrefix(action, "like") {
			twitter.LikeTweet(tweetLinks[index], r.UserID)
			sendTweetMessage(s, r.UserID, "✅ Successfully liked", tweetLinks[index])
		} else if strings.HasPrefix(action, "retweet") {
			// RetweetTweet(m.Content)
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
		tweetLinks := twitter.GetTweetLinks(m.Content)
		index := action[len(action)-1] - '0'
		if strings.HasPrefix(action, "like") {
			twitter.UnlikeTweet(tweetLinks[index], r.UserID)
			sendTweetMessage(s, r.UserID, "✅ Successfully unliked", tweetLinks[index])
		} else if strings.HasPrefix(action, "retweet") {
			// unretweetTweet(m.Content)
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
