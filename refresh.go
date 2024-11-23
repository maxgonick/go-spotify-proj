package main

import (
	"encoding/json"
	"os"
	"time"

	"golang.org/x/oauth2"
)

func GetRefresh() (oauth2.Token, bool) {
	haveValidRefresh := false
	refreshTokenFile, err := os.ReadFile(".refresh.json")
	check(err)
	var token oauth2.Token
	err = json.Unmarshal(refreshTokenFile, &token)
	check(err)
	currentTime := time.Now()
	expiryTime := token.Expiry
	check(err)
	if currentTime.Before(expiryTime) {
		haveValidRefresh = true
	}
	return token, haveValidRefresh
}
