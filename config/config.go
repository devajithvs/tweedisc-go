package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var Log = &logrus.Logger{
	Out:   os.Stderr,
	Hooks: make(logrus.LevelHooks),
	Level: logrus.DebugLevel,
}

var (
	DatabaseURL           string
	Token                 string
	ClientID              string
	ClientSecret          string
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	RedirectURI           string
	GumroadURI            string
	AccessToken           string
	Port                  string
	DatabasePrefix        string
)

func init() {
	var customFormatter = new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	Log.SetFormatter(customFormatter)

	log := Log
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	DatabaseURL = os.Getenv("DATABASE_URL")
	Token = os.Getenv("TOKEN")
	ClientID = os.Getenv("CLIENT_ID")
	ClientSecret = os.Getenv("CLIENT_SECRET")
	TwitterConsumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
	TwitterConsumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	RedirectURI = os.Getenv("REDIRECT_URI")
	GumroadURI = os.Getenv("GUMROAD_URI")
	AccessToken = os.Getenv("ACCESS_TOKEN")
	Port = os.Getenv("PORT")
	DatabasePrefix = "v2_"
}
