package notifier

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// SendTelegram sends a message and returns (success, errorMessage).
func SendTelegram(chatID, message string) (bool, string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Printf("[telegram] TELEGRAM_BOT_TOKEN not set, skipping message: %s", message)
		return false, "TELEGRAM_BOT_TOKEN not set"
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"Markdown"},
	})
	if err != nil {
		log.Printf("[telegram] failed to send message: %v", err)
		return false, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("status %d: %s", resp.StatusCode, string(body))
		log.Printf("[telegram] unexpected %s", errMsg)
		return false, errMsg
	}

	log.Printf("[telegram] message sent to %s", chatID)
	return true, ""
}
