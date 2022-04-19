package twitter

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"tweedisc-go/config"
	"tweedisc-go/database"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/like"
	"github.com/michimani/gotwi/tweet/like/types"
)

var log = config.Log

func GetTweetID(tweetLink string) string {
	texts := strings.Split(tweetLink, "status/")
	text := texts[len(texts)-1]

	temp := 0
	for _, chr := range text {
		if chr > '0' || chr < '9' {
			temp++
		} else {
			break
		}
	}
	log.Println("Tweet ID: ", text[:temp])
	return text[:temp]
}

func GetTweetAuthor(tweetLink string) string {
	firstHalf := strings.Split(tweetLink, "/status")
	middlePart := strings.Split(firstHalf[len(firstHalf)-2], "twitter.com/")
	return middlePart[len(middlePart)-1]
}

func IsTweet(link string) bool {
	if strings.Contains(link, "twitter.com/") && strings.Contains(link, "status") {
		return true
	}
	return false
}

func GetTweetLinks(content string) []string {
	r := regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
	urls := r.FindAllString(content, -1)
	tweets := urls[:0]
	for _, url := range urls {
		if IsTweet(url) {
			tweets = append(tweets, url)
		}
	}
	return tweets
}

func newOAuth1Client(user database.User) (*gotwi.Client, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user.Twitter_access_token,
		OAuthTokenSecret:     user.Twitter_access_token_secret,
	}

	return gotwi.NewClient(in)
}

func LikeTweet(tweetLink string, discordUserID string) error {

	user, err := database.GetUser(discordUserID)
	if err != nil {
		log.Warn(err)
		return err
	}

	c, err := newOAuth1Client(user)
	if err != nil {
		log.Warn(err)
		return err
	}

	tweetID := GetTweetID(tweetLink)
	p := &types.CreateInput{
		ID:      *gotwi.String(fmt.Sprint(user.Twitter_user_id)),
		TweetID: *gotwi.String(tweetID),
	}

	res, err := like.Create(context.Background(), c, p)
	if err != nil {
		log.Warn(err)
		return err
	}
	log.Println("Result: ", res)

	return nil
}

func UnlikeTweet(tweetLink string, discordUserID string) error {

	user, err := database.GetUser(discordUserID)
	if err != nil {
		log.Warn(err)
		return err
	}

	c, err := newOAuth1Client(user)
	if err != nil {
		log.Warn(err)
		return err
	}

	tweetID := GetTweetID(tweetLink)
	p := &types.DeleteInput{
		ID:      *gotwi.String(fmt.Sprint(user.Twitter_user_id)),
		TweetID: *gotwi.String(tweetID),
	}

	res, err := like.Delete(context.Background(), c, p)
	if err != nil {
		log.Warn(err)
		return err
	}
	log.Println("Result: ", res)

	return nil
}
