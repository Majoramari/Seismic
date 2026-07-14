package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/majoramari/seismic/apps/api/models"
)

type wakaTimeDay struct {
	Date       string                     `json:"date"`
	Heartbeats []models.WakaTimeHeartbeat `json:"heartbeats"`
}

type wakaTimeDataDump struct {
	Days []wakaTimeDay `json:"days"`
}

type dataDumpStatus struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	PercentComplete float64 `json:"percent_complete"`
	DownloadURL     string  `json:"download_url"`
}

const wakaTimeBaseURL = "https://wakatime.com/api/v1"

func wakaTimeAuthHeader(apiKey string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey))
}

// RequestWakaTimeHeartbeatDump kicks off an async export of the user's
// full heartbeat history. Returns the dump ID to poll for completion.
func RequestWakaTimeHeartbeatDump(apiKey string) (string, error) {
	body := []byte(`{"type":"heartbeats"}`)
	req, err := http.NewRequest("POST", wakaTimeBaseURL+"/users/current/data_dumps", newBytesReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", wakaTimeAuthHeader(apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wakatime rejected dump request (status %d): %s", resp.StatusCode, respBody)
	}

	var result struct {
		Data dataDumpStatus `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.ID, nil
}

// PollWakaTimeDump waits for the dump to finish, checking every 10
// seconds, up to a 20-minute ceiling. Returns the download URL once
// the export is ready.
func PollWakaTimeDump(ctx context.Context, apiKey, dumpID string) (string, error) {
	deadline := time.Now().Add(20 * time.Minute)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Second):
		}

		req, err := http.NewRequest("GET", wakaTimeBaseURL+"/users/current/data_dumps/"+dumpID, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", wakaTimeAuthHeader(apiKey))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		var result struct {
			Data dataDumpStatus `json:"data"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			return "", decodeErr
		}

		if result.Data.Status == "Completed" && result.Data.DownloadURL != "" {
			return result.Data.DownloadURL, nil
		}
	}

	return "", fmt.Errorf("timed out waiting for wakatime export after 20 minutes")
}

// DownloadWakaTimeDump fetches and parses the completed export.
func DownloadWakaTimeDump(downloadURL string) ([]models.WakaTimeHeartbeat, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dump wakaTimeDataDump
	if err := json.NewDecoder(resp.Body).Decode(&dump); err != nil {
		return nil, err
	}

	var all []models.WakaTimeHeartbeat
	for _, day := range dump.Days {
		all = append(all, day.Heartbeats...)
	}
	return all, nil
}

func newBytesReader(b []byte) io.Reader {
	return &byteReader{b: b}
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
