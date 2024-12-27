package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

func SendWebhook(payload interface{}, webhookURL string) error {
	client := http.Client{Timeout: constants.WebhookTimeout}
	parsedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(parsedPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook responded with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func SendWebhooks(
	ctx context.Context,
	payloads []interface{},
	getWebhookURL func(payload interface{}) string,
) []error {
	if len(payloads) == 0 {
		logger.GetLogger().Info("No payloads to send webhooks for.")
		return nil
	}

	// Limit concurrency with a semaphore
	sem := make(chan struct{}, constants.MaxWebhookWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, payload := range payloads {
		webhookURL := getWebhookURL(payload)
		if webhookURL == "" {
			logger.GetLogger().Warnf("No webhook URL provided for payload: %+v", payload)
			continue
		}

		sem <- struct{}{} // Acquire a semaphore slot
		wg.Add(1)

		go func(payload interface{}, webhookURL string) {
			defer func() {
				<-sem // Release the slot
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				logger.GetLogger().Warn("Context canceled before sending webhook.")
				return
			default:
				// Send the webhook
				if err := SendWebhook(payload, webhookURL); err != nil {
					logger.GetLogger().Errorf("Failed to send webhook for payload: %+v, error: %v", payload, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("payload: %+v, error: %w", payload, err))
					mu.Unlock()
				} else {
					logger.GetLogger().Infof("Webhook sent successfully for payload: %+v", payload)
				}
			}
		}(payload, webhookURL)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return errors
}
