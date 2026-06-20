// Package native provides zero-shell-exec Go implementations of common CLI tools.
// These replace shell_run calls with direct Go API calls, eliminating fork overhead
// (~20ms per call → ~0ms). For the most-used tools in the agent loop, this is a
// significant throughput improvement.
package native

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ── GitHub API (gh CLI replacement) ────────────────────────────────────

// GHClient wraps the GitHub REST API. Uses GITHUB_TOKEN or gh auth token.
type GHClient struct {
	Token  string
	Owner  string
	Repo   string
	client *http.Client
}

func NewGHClient(owner, repo string) *GHClient {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	return &GHClient{
		Token:  token,
		Owner:  owner,
		Repo:   repo,
		client: &http.Client{},
	}
}

// ── Pull Requests ─────────────────────────────────────────────────────

type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
	Body   string `json:"body"`
	Head   struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

// PRList returns open pull requests.
func (g *GHClient) PRList() ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open&per_page=30", g.Owner, g.Repo)
	body, err := g.get(url)
	if err != nil {
		return nil, err
	}
	var prs []PullRequest
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

// PRGet returns a single pull request by number.
func (g *GHClient) PRGet(number int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", g.Owner, g.Repo, number)
	body, err := g.get(url)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// PRCreate opens a new pull request.
func (g *GHClient) PRCreate(title, body, head, base string) (*PullRequest, error) {
	payload := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}
	payloadBytes, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", g.Owner, g.Repo)
	respBody, err := g.post(url, payloadBytes)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	if err := json.Unmarshal(respBody, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// PRFiles returns the list of files changed in a PR.
func (g *GHClient) PRFiles(number int) ([]PRFile, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", g.Owner, g.Repo, number)
	body, err := g.get(url)
	if err != nil {
		return nil, err
	}
	var files []PRFile
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, err
	}
	return files, nil
}

type PRFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
	Patch     string `json:"patch"`
}

// PRDiff returns the diff for a PR.
func (g *GHClient) PRDiff(number int) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", g.Owner, g.Repo, number)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("Authorization", "Bearer "+g.Token)
	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	diff, _ := io.ReadAll(resp.Body)
	return string(diff), nil
}

// ── Issues ────────────────────────────────────────────────────────────

type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
	URL    string `json:"html_url"`
}

// IssueGet returns an issue by number.
func (g *GHClient) IssueGet(number int) (*Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", g.Owner, g.Repo, number)
	body, err := g.get(url)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// ── Search ────────────────────────────────────────────────────────────

// SearchCode searches the repository code via GitHub API.
func (g *GHClient) SearchCode(query string) ([]CodeResult, error) {
	url := fmt.Sprintf("https://api.github.com/search/code?q=%s+repo:%s/%s&per_page=30",
		query, g.Owner, g.Repo)
	body, err := g.get(url)
	if err != nil {
		return nil, err
	}
	var result struct {
		Items []CodeResult `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

type CodeResult struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	URL     string `json:"html_url"`
	Repo    string `json:"repository.full_name"`
}

// ── HTTP helpers ──────────────────────────────────────────────────────

func (g *GHClient) get(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gh API %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}

func (g *GHClient) post(url string, payload []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gh API %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}
