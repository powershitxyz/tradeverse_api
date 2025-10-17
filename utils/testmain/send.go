package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func tokenSource(ctx context.Context) oauth2.TokenSource {
	GMAIL_CREDENTIAL := os.Getenv("GMAIL_CREDENTIAL")
	cred, err := os.ReadFile(GMAIL_CREDENTIAL)
	if err != nil {
		log.Fatalf("read credentials.json: %v", err)
	}
	cfg, err := google.ConfigFromJSON(cred, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("parse credentials: %v", err)
	}

	// 尝试加载本地 token
	if b, err := os.ReadFile("token.json"); err == nil {
		var t oauth2.Token
		if err := json.Unmarshal(b, &t); err == nil {
			return cfg.TokenSource(ctx, &t)
		}
	}

	// 首次授权
	authURL := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Println("Open this URL, authorize, then paste the code:\n", authURL)
	fmt.Print("Code: ")
	var code string
	fmt.Scan(&code)
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("exchange: %v", err)
	}
	if b, err := json.Marshal(tok); err == nil {
		_ = os.WriteFile("token.json", b, 0600)
	}
	return cfg.TokenSource(ctx, tok)
}

func sendHTML(from, to, subject, html string) error {
	ctx := context.Background()
	ts := tokenSource(ctx)
	svc, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return err
	}

	raw := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, html,
	)
	msg := &gmail.Message{Raw: base64.URLEncoding.EncodeToString([]byte(raw))}
	_, err = svc.Users.Messages.Send("me", msg).Do()
	return err
}

func main() {
	_ = godotenv.Load(".env")
	from := os.Getenv("GMAIL_SENDER") // 你的 Gmail
	to := os.Getenv("GMAIL_TEST_TO")  // 收件人
	if from == "" || to == "" {
		log.Fatal("set GMAIL_SENDER and GMAIL_TEST_TO")
	}

	html := `<div style="font-family:sans-serif">Hello from <b>Gmail API</b>.</div>`
	if err := sendHTML(from, to, "Gmail API test", html); err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent")
}
