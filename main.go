package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/nekomeowww/exchange-smtp-client/email"
)

type microsoftGraphOauth struct {
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

func obtainOauthAccessToken() (string, error) {
	encoded := url.Values{}
	encoded.Set("grant_type", "password")
	encoded.Set("username", os.Getenv("SMTP_EMAIL"))
	encoded.Set("password", os.Getenv("SMTP_PASSWORD"))
	encoded.Set("client_id", os.Getenv("SMTP_CLIENT_ID"))
	encoded.Set("client_secret", os.Getenv("SMTP_CLIENT_SECRET"))
	encoded.Set("scope", "https://outlook.office.com/SMTP.Send https://graph.microsoft.com/User.Read email openid offline_access")
	request, err := http.NewRequest(http.MethodPost, "https://login.microsoftonline.com/organizations/oauth2/v2.0/token", strings.NewReader(encoded.Encode()))

	client := http.Client{}
	respRaw, err := client.Do(request)
	if err != nil {
		return "", err
	}
	respBytes, err := ioutil.ReadAll(respRaw.Body)
	if err != nil {
		return "", err
	}

	var oauthResp microsoftGraphOauth
	err = json.Unmarshal(respBytes, &oauthResp)
	if err != nil {
		return "", nil
	}

	return oauthResp.AccessToken, nil
}

func main() {
	senderAddress := "sender@outlook.com"
	smtpHost := "smtp.office365.com"
	smtpPort := 587

	accessToken, err := obtainOauthAccessToken()
	if err != nil {
		log.Fatal(err)
		return
	}

	sender, err := email.NewMessageSender(smtpHost, smtpPort, senderAddress, accessToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	message := email.NewMessage()
	message.To([]string{"test@gmail.com"})
	message.Subject("test")
	message.Body("text/plain", "Test Mail")
	message.Attach([]byte{}, "filename.ext") // attach file if wished

	err = sender.Send(message)
	if err != nil {
		fmt.Println(err)
		return
	}
}
