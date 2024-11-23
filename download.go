package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func Download(client *http.Client, ctx context.Context) bool {
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
	cmd := exec.Command("yt-dlp", "-o", "music/%(playlist_index)s - %(title)s.%(ext)s", playlist.Id)
	cmd.Stdout = os.Stdout
	cmd.Run()
	return true
}
