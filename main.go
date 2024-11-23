package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type Client_Secret struct {
	Installed struct {
		Client_id                   string   `json:"client_id"`
		Project_id                  string   `json:"project_id"`
		Auth_uri                    string   `json:"auth_uri"`
		Token_uri                   string   `json:"token_uri"`
		Auth_provider_x509_cert_url string   `json:"auth_provider_x509_cert_url"`
		Client_secret               string   `json:"client_secret"`
		Redirect_uris               []string `json:"redirect_uris"`
	} `json:"installed"`
}

type Refresh_Token struct {
	RefreshToken string `json:"refreshToken"`
	Expiry       string `json:"expiry"`
}

type Refresh_File struct {
	Tokens []Refresh_Token `json:"tokens"`
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {

	jsonData, err := os.ReadFile("client_secret.json")
	if err != nil {
		panic(err)
	}

	var client_secret Client_Secret
	err = json.Unmarshal(jsonData, &client_secret)
	check(err)

	token, validToken := GetRefresh()

	ctx := context.Background()
	var conf = &oauth2.Config{
		ClientID:     client_secret.Installed.Client_id,
		ClientSecret: client_secret.Installed.Client_secret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeScope},
		RedirectURL:  "http://localhost:8080/",
	}

	//Use Refresh Token instead of re-authenticating
	if validToken {
		client := conf.Client(ctx, &token)
		result := Download(client, ctx)
		if result {
			fmt.Println("Download Completed!")
		} else {
			fmt.Println("Download Failed :(")
		}
	} else {
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		verifier := oauth2.GenerateVerifier()
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
		fmt.Printf("Visit the URL for the auth dialog: %v", url)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")

			if code != "" {
				tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
				check(err)
				tokenJSON, err := json.Marshal(tok)
				check(err)
				os.WriteFile(".refresh.json", tokenJSON, 0666)
				fmt.Printf("%s\n", tokenJSON)
				client := conf.Client(ctx, tok)
				result := Download(client, ctx)
				if result {
					fmt.Println("Download Completed!")
				} else {
					fmt.Println("Download Failed :(")
				}
				os.Exit(0)
			}

		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}
