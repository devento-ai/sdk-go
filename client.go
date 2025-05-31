package tavor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.tavor.dev"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		if debug {
			c.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
		}
	}
}

func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("TAVOR_API_KEY")
	}
	if apiKey == "" {
		return nil, NewAuthenticationError("API key is required. Pass it as a parameter or set TAVOR_API_KEY environment variable")
	}

	baseURL := os.Getenv("TAVOR_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client := &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)), // no-op logger by default
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (c *Client) CreateBox(ctx context.Context, config *BoxConfig) (*BoxHandle, error) {
	if config == nil {
		config = &BoxConfig{}
	}

	if config.Template == "" && config.TemplateID == "" {
		if envTemplate := os.Getenv("TAVOR_BOX_TEMPLATE"); envTemplate != "" {
			config.Template = BoxTemplate(envTemplate)
		}
	}

	if config.Timeout == 0 {
		if envTimeout := os.Getenv("TAVOR_BOX_TIMEOUT"); envTimeout != "" {
			if timeout, err := strconv.Atoi(envTimeout); err == nil {
				config.Timeout = timeout
			}
		}
	}

	req := createBoxRequest{
		Metadata: config.Metadata,
	}

	if config.TemplateID != "" {
		req.BoxTemplate = config.TemplateID
	} else if config.Template != "" {
		switch config.Template {
		case BoxTemplateBasic:
			req.BoxTemplate = "Basic"
		case BoxTemplatePro:
			req.BoxTemplate = "Pro"
		}
	} else {
		// default to Basic template if none specified
		req.BoxTemplate = "Basic"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("creating box", "config", string(body))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/boxes", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleError(resp)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("box created", "response", string(body))

	var boxResp createBoxResponse
	if err := json.Unmarshal(body, &boxResp); err != nil {
		return nil, err
	}

	// Create a minimal Box object with just the ID, since that's what
	// the API responds with
	box := &Box{
		ID:     boxResp.ID,
		Status: BoxStatusQueued,
	}

	return newBoxHandle(c, box), nil
}

func (c *Client) ListBoxes(ctx context.Context) ([]*Box, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/boxes", nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var listResp listBoxesResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	boxes := make([]*Box, len(listResp.Data))
	for i := range listResp.Data {
		boxes[i] = &listResp.Data[i]
	}

	return boxes, nil
}

func (c *Client) GetBox(ctx context.Context, boxID string) (*BoxHandle, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/boxes/"+boxID, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var boxResp getBoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxResp); err != nil {
		return nil, err
	}

	return newBoxHandle(c, &boxResp.Data), nil
}

func (c *Client) WithSandbox(ctx context.Context, fn func(context.Context, *BoxHandle) error, config *BoxConfig) error {
	box, err := c.CreateBox(ctx, config)
	if err != nil {
		return err
	}

	c.logger.Debug("created box in WithSandbox", "boxID", box.ID())

	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := box.Stop(stopCtx); err != nil {
			c.logger.Error("failed to stop box", "boxID", box.ID(), "error", err)
		}
	}()

	if err := box.WaitUntilReady(ctx); err != nil {
		return err
	}

	return fn(ctx, box)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("User-Agent", "tavor-go-sdk/"+Version)
}

func (c *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewAPIError(resp.StatusCode, "Failed to read error response")
	}

	c.logger.Debug("API error response", "statusCode", resp.StatusCode, "body", string(body))

	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return NewAPIError(resp.StatusCode, "API generic error:"+string(body))
	}

	return parseError(resp.StatusCode, &errResp)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyBytes)

		c.logger.Debug("making request", "method", method, "path", path, "body", string(bodyBytes))
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}
