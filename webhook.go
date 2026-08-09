package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookManager struct {
	webhooks []Webhook
	client   *http.Client
}

type Webhook struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

type WebhookPayload struct {
	Event     string    `json:"event"`
	Task      Task      `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}

func NewWebhookManager() *WebhookManager {
	return &WebhookManager{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *WebhookManager) AddWebhook(url string, events []string) {
	m.webhooks = append(m.webhooks, Webhook{URL: url, Events: events, Active: true})
}

func (m *WebhookManager) RemoveWebhook(url string) {
	for i, wh := range m.webhooks {
		if wh.URL == url {
			m.webhooks = append(m.webhooks[:i], m.webhooks[i+1:]...)
			return
		}
	}
}

func (m *WebhookManager) SendWebhook(event string, task Task) {
	payload := WebhookPayload{Event: event, Task: task, Timestamp: time.Now()}
	data, _ := json.Marshal(payload)
	for _, wh := range m.webhooks {
		if !wh.Active {
			continue
		}
		for _, e := range wh.Events {
			if e == event || e == "*" {
				go m.send(wh.URL, data)
				break
			}
		}
	}
}

func (m *WebhookManager) send(url string, data []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GoTaskCLI/1.0")
	resp, err := m.client.Do(req)
	if err != nil {
		fmt.Printf("Webhook error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

func (m *WebhookManager) GetWebhooks() []Webhook {
	return m.webhooks
}

func (m *WebhookManager) TestWebhook(url string) error {
	payload := map[string]interface{}{
		"event":     "test",
		"message":   "Webhook test from Go Task CLI",
		"timestamp": time.Now(),
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
