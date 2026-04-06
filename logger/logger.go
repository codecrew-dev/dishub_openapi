package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type LogField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type SystemLogPayload struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Color       int        `json:"color,omitempty"`
	URL         string     `json:"url,omitempty"`
	Fields      []LogField `json:"fields,omitempty"`
}

func SendSystemLog(payload SystemLogPayload) {
	go func() {
		ipcURL := os.Getenv("BOT_IPC_URL")
		syncToken := os.Getenv("BOT_SYNC_TOKEN")

		if ipcURL == "" || syncToken == "" {
			log.Println("[Logger] BOT_IPC_URL or BOT_SYNC_TOKEN not set. Skipping log.")
			return
		}

		ipcURL = strings.TrimRight(ipcURL, "/") + "/log_action"
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Logger] Error marshalling payload: %v\n", err)
			return
		}

		req, err := http.NewRequest("POST", ipcURL, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Logger] Error creating request: %v\n", err)
			return
		}

		req.Header.Set("Authorization", syncToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Logger] Error sending log request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Logger] IPC returned non-200 status code: %d\n", resp.StatusCode)
		}
	}()
}
