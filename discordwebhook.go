package discordwebhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func SendMessage(url string, message Message, proxy *url.URL) error {
	// Validate parameters
	if url == "" {
		return errors.New("empty Webhook URL")
	}

	client := &http.Client{}

	if proxy != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	for {
		payload := new(bytes.Buffer)

		err := json.NewEncoder(payload).Encode(message)
		if err != nil {
			return err
		}

		request, err := http.NewRequest("POST", url, payload)
		if err != nil {
			return err
		}

		request.Header.Set("Content-Type", "application/json")

		// Make the HTTP request
		resp, err := client.Do(request)

		if err != nil {
			return err
		}

		switch resp.StatusCode {
		case http.StatusOK, http.StatusNoContent:
			// Success
			_ = resp.Body.Close()
			return nil
		case http.StatusTooManyRequests:
			// Rate limit exceeded, retry after backoff duration
			resetAfter := resp.Header.Get("Retry-After")
			parsedAfter, err := strconv.ParseFloat(resetAfter, 64)
			if err != nil {
				return err
			}

			whole, frac := math.Modf(parsedAfter)
			resetAt := time.Now().Add(time.Duration(whole) * time.Second).Add(time.Duration(frac*1000) * time.Millisecond).Add(250 * time.Millisecond)

			time.Sleep(time.Until(resetAt))

		default:
			// Handle other HTTP status codes
			_ = resp.Body.Close()
			return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
		}
	}
}
