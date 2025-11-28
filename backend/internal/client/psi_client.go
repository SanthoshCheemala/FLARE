// Package client implements the PSI client.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SanthoshCheemala/FLARE/backend/internal/models"
	"github.com/SanthoshCheemala/FLARE/backend/internal/psiadapter"
)

type PSIClient struct {
	serverURL string
	client    *http.Client
}

func NewPSIClient(serverURL string) *PSIClient {
	return &PSIClient{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for PSI operations
		},
	}
}

type InitSessionRequest struct {
	SanctionListIDs []string `json:"sanctionListIds"`
}

type InitSessionResponse struct {
	SessionID string                             `json:"sessionId"`
	Params    *psiadapter.SerializedServerParams `json:"params"`
}

func (c *PSIClient) InitSession(ctx context.Context, sanctionListIDs []string) (string, *psiadapter.SerializedServerParams, error) {
	reqBody := InitSessionRequest{
		SanctionListIDs: sanctionListIDs,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/session/init", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var initResp InitSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return initResp.SessionID, initResp.Params, nil
}

type IntersectRequest struct {
	SessionID   string                        `json:"sessionId"`
	Ciphertexts []psiadapter.ClientCiphertext `json:"ciphertexts"`
}

type IntersectResponse struct {
	Matches []uint64 `json:"matches"`
}

func (c *PSIClient) Intersect(ctx context.Context, sessionID string, ciphertexts []psiadapter.ClientCiphertext) ([]uint64, error) {
	reqBody := IntersectRequest{
		SessionID:   sessionID,
		Ciphertexts: ciphertexts,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/session/intersect", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var intersectResp IntersectResponse
	if err := json.NewDecoder(resp.Body).Decode(&intersectResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return intersectResp.Matches, nil
}

type SanctionList struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	RecordCount int    `json:"record_count"`
	CreatedAt   string `json:"created_at"`
}

func (c *PSIClient) GetSanctionLists(ctx context.Context) ([]SanctionList, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/lists/sanctions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var lists []SanctionList
	if err := json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return lists, nil
}

// ResolveSanctions fetches full sanction details for matched hashes from the Server
func (c *PSIClient) ResolveSanctions(ctx context.Context, sessionID string, hashes []uint64) ([]*models.Sanction, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"hashes": hashes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/session/%s/resolve", c.serverURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Sanctions []struct {
			Hash    uint64 `json:"hash"`
			Name    string `json:"name"`
			DOB     string `json:"dob"`
			Country string `json:"country"`
			Program string `json:"program"`
			Source  string `json:"source"`
		} `json:"sanctions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to Sanction models
	sanctions := make([]*models.Sanction, len(result.Sanctions))
	for i, s := range result.Sanctions {
		sanctions[i] = &models.Sanction{
			Hash:    s.Hash,
			Name:    s.Name,
			DOB:     s.DOB,
			Country: s.Country,
			Program: s.Program,
			Source:  s.Source,
			ListID:  0, // These are fetched from remote, no local list ID
		}
	}

	return sanctions, nil
}
