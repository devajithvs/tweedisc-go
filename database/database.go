package database

import (
	"context"
	"path/filepath"
	"tweedisc-go/config"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

var (
	database *db.Client
	ok       error
)

func init() {
	serviceAccountKeyFilePath, err := filepath.Abs("./serviceAccountKey.json")
	if err != nil {
		panic("Unable to load serviceAccountKeys.json file")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath) //Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Firebase load error")
	}

	database, ok = app.DatabaseWithURL(context.Background(), config.DatabaseURL)
	if ok != nil {
		panic(ok)
	}
}

type Server struct {
	guild_id           int
	guild_name         string
	gumroad_url        string
	updated_time_stamp int
}

func GetServer(guildID string) (Server, error) {
	var result Server
	ref := database.NewRef("/" + config.DatabasePrefix + "tweedisc/servers")
	err := ref.Child(guildID).Get(context.Background(), &result)
	return result, err
}

type User struct {
	Created_time_stamp          int
	Discord_access_token        string
	Discord_code                string
	Discord_expiry              int
	Discord_scope               string
	Discord_user_id             string
	Discord_user_name           string
	Secret                      string
	State                       string
	Twitter_access_token        string
	Twitter_access_token_secret string
	Twitter_oauth_token         string
	Twitter_oauth_verifier      string
	Twitter_user_id             int
	Twitter_user_name           string
	Updated_time_stamp          float32
}

func GetUser(discordUserID string) (User, error) {
	var result User
	ref := database.NewRef("/" + config.DatabasePrefix + "tweedisc/users")
	err := ref.Child(discordUserID).Get(context.Background(), &result)
	return result, err
}
