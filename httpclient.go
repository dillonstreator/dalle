package dalle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	libraryVersion   = "1.0.0"
	defaultUserAgent = "dalle/" + libraryVersion
	baseURL          = "https://labs.openai.com/api/labs"

	defaultHTTPClientTimeout = 5 * time.Second
)

type option func(*HTTPClient) error

func WithHTTPClient(httpClient *http.Client) option {
	return func(c *HTTPClient) error {
		c.httpClient = httpClient

		return nil
	}
}

func WithUserAgent(userAgent string) option {
	return func(c *HTTPClient) error {
		c.userAgent = userAgent

		return nil
	}
}

type HTTPClient struct {
	httpClient *http.Client
	userAgent  string
	apiKey     string
}

var _ Client = (*HTTPClient)(nil)

func NewHTTPClient(apiKey string, opts ...option) (*HTTPClient, error) {
	c := &HTTPClient{
		httpClient: &http.Client{Timeout: defaultHTTPClientTimeout},
		userAgent:  defaultUserAgent,
		apiKey:     apiKey,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

type Task struct {
	Object      string      `json:"object"`
	ID          string      `json:"id"`
	Created     int64       `json:"created"`
	TaskType    string      `json:"task_type"`
	Status      string      `json:"status"`
	PromptID    string      `json:"prompt_id"`
	Prompt      Prompt      `json:"prompt"`
	Generations Generations `json:"generations"`
}
type Generations struct {
	Data   []GenerationData `json:"data"`
	Object string           `json:"object"`
}
type GenerationData struct {
	Created        int64      `json:"created"`
	Generation     Generation `json:"generation"`
	GenerationType string     `json:"generation_type"`
	ID             string     `json:"id"`
}
type Generation struct {
	ImagePath string `json:"image_path"`
}

type Prompt struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	PromptType string `json:"prompt_type"`
	Prompt     struct {
		Caption string `json:"caption"`
	} `json:"prompt"`
	ParentGenerationID string `json:"parent_generation_id"`
}

type GenerateRequest struct {
	Prompt   GenerateRequestPrompt `json:"prompt"`
	TaskType string                `json:"task_type"`
}
type GenerateRequestPrompt struct {
	BatchSize int32  `json:"batch_size"`
	Caption   string `json:"caption"`
}

func (c *HTTPClient) Generate(ctx context.Context, caption string) (*Task, error) {
	task := &Task{}
	req := &GenerateRequest{
		Prompt: GenerateRequestPrompt{
			BatchSize: 4,
			Caption:   caption,
		},
		TaskType: "text2im",
	}
	return task, c.request(ctx, "POST", "/tasks", req, task)
}

func (c *HTTPClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	task := &Task{}
	return task, c.request(ctx, "GET", "/tasks/"+taskID, nil, task)
}

func (c *HTTPClient) Download(ctx context.Context, generationID string) (io.ReadCloser, error) {
	url := baseURL + "/generations/" + generationID + "/download"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request %w", err)
	}

	return resp.Body, nil
}

func (c *HTTPClient) request(ctx context.Context, method, path string, data interface{}, result interface{}) error {
	url := baseURL + path

	var body io.Reader
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("parsing request data %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("building request %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// TODO: improve error handling...
	if resp.StatusCode != http.StatusOK {
		return Error{
			Message:    "unexpected non 200 status code",
			StatusCode: resp.StatusCode,
			Details:    string(respBody),
		}
	}

	if err = json.Unmarshal(respBody, result); err != nil {
		return Error{
			Message:    err.Error(),
			StatusCode: resp.StatusCode,
			Details:    string(respBody),
		}
	}

	return nil
}

type Error struct {
	Message    string
	StatusCode int
	Details    string
}

func (e Error) Error() string {
	return fmt.Sprintf("dalle: %s (status: %d, details: %s)", e.Message, e.StatusCode, e.Details)
}
