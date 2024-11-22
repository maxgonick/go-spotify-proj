package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
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

func main() {

	jsonData, err := os.ReadFile("client_secret.json")
	if err != nil {
		panic(err)
	}

	var client_secret Client_Secret
	err = json.Unmarshal(jsonData, &client_secret)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	var conf = &oauth2.Config{
		ClientID:     client_secret.Installed.Client_id,
		ClientSecret: client_secret.Installed.Client_secret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeScope},
		RedirectURL:  "http://localhost:8080/",
	}

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
			if err != nil {
				log.Fatal(err)
			}

			client := conf.Client(ctx, tok)

			youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
			if err != nil {
				log.Fatal(err)
			}
			playListsCall := youtubeService.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true)
			playListsResp, err := playListsCall.Do()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Playlists:")
			for index, playlist := range playListsResp.Items {
				fmt.Printf("[%d] Title: %v, ID: %v\n", index, playlist.Snippet.Title, playlist.Id)
			}

			fmt.Println("Select a playlist:")
			var playListIndex int
			fmt.Scanln(&playListIndex)
			playlist := playListsResp.Items[playListIndex]
			// TODO: Error Handling
			fmt.Println("Selected playlist: ", playlist.Id)

			cmd := exec.Command("yt-dlp", "-o", "~/dev/go-spotify-proj/music/%(playlist)s/%(playlist_index)s - %(title)s.%(ext)s", playlist.Id)

			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))
		}

	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// yt-dl -o '%(playlist)s/%(playlist_index)s - %(title)s.%(ext)s'
