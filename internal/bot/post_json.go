package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (b *Bot) postJSON(ctx context.Context, method string, v any) error {
	if b.postJSONFunc != nil {
		return b.postJSONFunc(ctx, method, v)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.telegram.org/bot"+b.token+"/"+method,
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tg %s: %s %s", method, resp.Status, string(body))
	}

	var r struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if !r.OK {
		return fmt.Errorf("tg %s: not ok", method)
	}

	return nil
}

func (b *Bot) getUpdates(ctx context.Context, offset int) ([]tgUpdate, error) {
	body := map[string]any{
		"offset":  offset,
		"timeout": 30,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.telegram.org/bot"+b.token+"/getUpdates",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tg getUpdates: %s %s", resp.Status, string(raw))
	}

	var r struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if !r.OK {
		return nil, fmt.Errorf("tg getUpdates: not ok")
	}

	return r.Result, nil
}
