package infrastructure

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

const (
	githubBugReportAPIBaseURL  = "https://api.github.com"
	githubBugReportAPIVersion  = "2026-03-10"
	githubBugReportMaxResponse = 4 * 1024 * 1024
)

type GitHubBugReportClientConfig struct {
	APIBaseURL string
	Token      string
	Repository string
	HTTPClient *http.Client
}

type GitHubBugReportIssueClient struct {
	client     *githubBugReportClient
	repository githubBugReportRepository
}

type GitHubBugReportAssetClient struct {
	client     *githubBugReportClient
	repository githubBugReportRepository
}

type githubBugReportClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type githubBugReportRepository struct {
	Owner string
	Name  string
}

type GitHubBugReportError struct {
	Operation  string
	StatusCode int
}

func (err *GitHubBugReportError) Error() string {
	if err.StatusCode > 0 {
		return fmt.Sprintf("GitHub bug report %s failed with status %d", err.Operation, err.StatusCode)
	}
	return fmt.Sprintf("GitHub bug report %s failed", err.Operation)
}

func NewGitHubBugReportIssueClient(config GitHubBugReportClientConfig) (*GitHubBugReportIssueClient, error) {
	client, repository, err := newGitHubBugReportClient(config)
	if err != nil {
		return nil, err
	}
	return &GitHubBugReportIssueClient{client: client, repository: repository}, nil
}

func NewGitHubBugReportAssetClient(config GitHubBugReportClientConfig) (*GitHubBugReportAssetClient, error) {
	client, repository, err := newGitHubBugReportClient(config)
	if err != nil {
		return nil, err
	}
	return &GitHubBugReportAssetClient{client: client, repository: repository}, nil
}

func newGitHubBugReportClient(
	config GitHubBugReportClientConfig,
) (*githubBugReportClient, githubBugReportRepository, error) {
	repository, err := parseGitHubBugReportRepository(config.Repository)
	if err != nil {
		return nil, githubBugReportRepository{}, err
	}
	if strings.TrimSpace(config.Token) == "" {
		return nil, githubBugReportRepository{}, errors.New("GitHub bug report token is required")
	}
	baseURL := strings.TrimRight(config.APIBaseURL, "/")
	if baseURL == "" {
		baseURL = githubBugReportAPIBaseURL
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &githubBugReportClient{baseURL: baseURL, token: config.Token, httpClient: httpClient}, repository, nil
}

func ValidateGitHubBugReportRepositoryTargets(issueRepository string, assetRepository string) error {
	issue, err := parseGitHubBugReportRepository(issueRepository)
	if err != nil {
		return fmt.Errorf("parse issue repository: %w", err)
	}
	asset, err := parseGitHubBugReportRepository(assetRepository)
	if err != nil {
		return fmt.Errorf("parse asset repository: %w", err)
	}
	if strings.EqualFold(issue.Owner+"/"+issue.Name, asset.Owner+"/"+asset.Name) {
		return errors.New("GitHub bug report repositories must differ")
	}
	return nil
}

func parseGitHubBugReportRepository(value string) (githubBugReportRepository, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 || !validGitHubBugReportName(parts[0]) || !validGitHubBugReportName(parts[1]) {
		return githubBugReportRepository{}, errors.New("GitHub repository must use owner/name")
	}
	return githubBugReportRepository{Owner: parts[0], Name: parts[1]}, nil
}

func validGitHubBugReportName(value string) bool {
	if value == "" || len(value) > 100 || value == "." || value == ".." {
		return false
	}
	for _, valueRune := range value {
		if (valueRune >= 'a' && valueRune <= 'z') || (valueRune >= 'A' && valueRune <= 'Z') ||
			(valueRune >= '0' && valueRune <= '9') || valueRune == '-' || valueRune == '_' || valueRune == '.' {
			continue
		}
		return false
	}
	return true
}

func (client *GitHubBugReportIssueClient) CreateIssue(
	ctx context.Context,
	issue application.BugReportIssue,
) (application.GitHubBugReportIssue, error) {
	payload := struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels"`
	}{Title: issue.Title, Body: issue.Body, Labels: issue.Labels}
	var response struct {
		Number int64  `json:"number"`
		URL    string `json:"html_url"`
		Body   string `json:"body"`
	}
	if err := client.client.requestJSON(ctx, http.MethodPost, client.repository.apiPath("issues"), nil,
		payload, &response, http.StatusCreated, http.StatusOK); err != nil {
		return application.GitHubBugReportIssue{}, err
	}
	return application.GitHubBugReportIssue{Number: response.Number, URL: response.URL, Body: response.Body}, nil
}

func (client *GitHubBugReportIssueClient) ListRecentIssues(
	ctx context.Context,
	since time.Time,
	limit int,
) ([]application.GitHubBugReportIssue, error) {
	limit = min(max(limit, 1), 1_000)
	type issueResponse struct {
		Number      int64          `json:"number"`
		URL         string         `json:"html_url"`
		Body        string         `json:"body"`
		PullRequest map[string]any `json:"pull_request"`
	}
	issues := make([]application.GitHubBugReportIssue, 0, min(limit, 100))
	examined := 0
	for page := 1; examined < limit; page++ {
		pageSize := min(100, limit-examined)
		query := url.Values{
			"state": {"all"}, "since": {since.UTC().Format(time.RFC3339)},
			"per_page": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(page)},
			"sort": {"created"}, "direction": {"desc"},
		}
		var response []issueResponse
		if err := client.client.requestJSON(ctx, http.MethodGet, client.repository.apiPath("issues"), query,
			nil, &response, http.StatusOK); err != nil {
			return nil, err
		}
		examined += len(response)
		for _, item := range response {
			if item.PullRequest != nil {
				continue
			}
			issues = append(issues, application.GitHubBugReportIssue{
				Number: item.Number, URL: item.URL, Body: item.Body,
			})
		}
		if len(response) < pageSize {
			break
		}
	}
	return issues, nil
}

func (client *GitHubBugReportIssueClient) ListIssueComments(
	ctx context.Context,
	issueNumber int64,
	limit int,
) ([]application.GitHubBugReportComment, error) {
	limit = min(max(limit, 1), 1_000)
	type commentResponse struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}
	endpoint := client.repository.apiPath("issues", strconv.FormatInt(issueNumber, 10), "comments")
	comments := make([]application.GitHubBugReportComment, 0, min(limit, 100))
	examined := 0
	for page := 1; examined < limit; page++ {
		pageSize := min(100, limit-examined)
		query := url.Values{"per_page": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(page)}}
		var response []commentResponse
		if err := client.client.requestJSON(ctx, http.MethodGet, endpoint, query, nil, &response, http.StatusOK); err != nil {
			return nil, err
		}
		examined += len(response)
		for _, comment := range response {
			comments = append(comments, application.GitHubBugReportComment{ID: comment.ID, Body: comment.Body})
		}
		if len(response) < pageSize {
			break
		}
	}
	return comments, nil
}

func (client *GitHubBugReportIssueClient) CreateIssueComment(
	ctx context.Context,
	issueNumber int64,
	body string,
) (application.GitHubBugReportComment, error) {
	var response struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}
	endpoint := client.repository.apiPath("issues", strconv.FormatInt(issueNumber, 10), "comments")
	if err := client.client.requestJSON(ctx, http.MethodPost, endpoint, nil, map[string]string{"body": body},
		&response, http.StatusCreated, http.StatusOK); err != nil {
		return application.GitHubBugReportComment{}, err
	}
	return application.GitHubBugReportComment{ID: response.ID, Body: response.Body}, nil
}

func (client *GitHubBugReportAssetClient) CreateFile(
	ctx context.Context,
	filePath string,
	message string,
	content []byte,
) (application.GitHubBugReportAsset, error) {
	cleanPath, err := validGitHubBugReportAssetPath(filePath)
	if err != nil {
		return application.GitHubBugReportAsset{}, err
	}
	payload := struct {
		Message string `json:"message"`
		Content string `json:"content"`
	}{Message: message, Content: base64.StdEncoding.EncodeToString(content)}
	var response struct {
		Content struct {
			SHA string `json:"sha"`
		} `json:"content"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	endpoint := client.repository.apiPath("contents") + "/" + escapeGitHubBugReportPath(cleanPath)
	if err := client.client.requestJSON(ctx, http.MethodPut, endpoint, nil, payload, &response,
		http.StatusCreated, http.StatusOK); err != nil {
		return application.GitHubBugReportAsset{}, err
	}
	return client.asset(cleanPath, response.Content.SHA, response.Commit.SHA), nil
}

func (client *GitHubBugReportAssetClient) FindFile(
	ctx context.Context,
	filePath string,
) (bool, application.GitHubBugReportAsset, error) {
	cleanPath, err := validGitHubBugReportAssetPath(filePath)
	if err != nil {
		return false, application.GitHubBugReportAsset{}, err
	}
	var contentResponse struct {
		SHA string `json:"sha"`
	}
	endpoint := client.repository.apiPath("contents") + "/" + escapeGitHubBugReportPath(cleanPath)
	err = client.client.requestJSON(ctx, http.MethodGet, endpoint, nil, nil, &contentResponse,
		http.StatusOK, http.StatusNotFound)
	if githubError, ok := err.(*GitHubBugReportError); ok && githubError.StatusCode == http.StatusNotFound {
		return false, application.GitHubBugReportAsset{}, nil
	}
	if err != nil {
		return false, application.GitHubBugReportAsset{}, err
	}
	query := url.Values{"path": {cleanPath}, "per_page": {"1"}}
	var commits []struct {
		SHA string `json:"sha"`
	}
	if err := client.client.requestJSON(ctx, http.MethodGet, client.repository.apiPath("commits"), query,
		nil, &commits, http.StatusOK); err != nil {
		return false, application.GitHubBugReportAsset{}, err
	}
	if len(commits) == 0 || commits[0].SHA == "" {
		return false, application.GitHubBugReportAsset{}, &GitHubBugReportError{Operation: "find asset commit"}
	}
	return true, client.asset(cleanPath, contentResponse.SHA, commits[0].SHA), nil
}

func (client *GitHubBugReportAssetClient) asset(
	filePath string,
	blobSHA string,
	commitSHA string,
) application.GitHubBugReportAsset {
	assetURL := "https://raw.githubusercontent.com/" + url.PathEscape(client.repository.Owner) + "/" +
		url.PathEscape(client.repository.Name) + "/" + url.PathEscape(commitSHA) + "/" + escapeGitHubBugReportPath(filePath)
	return application.GitHubBugReportAsset{BlobSHA: blobSHA, CommitSHA: commitSHA, URL: assetURL}
}

func ValidateGitHubBugReportCapability(
	ctx context.Context,
	issueClient *GitHubBugReportIssueClient,
	assetClient *GitHubBugReportAssetClient,
	requiredLabels []string,
) error {
	if issueClient == nil {
		return errors.New("GitHub bug report clients are required")
	}
	if _, err := issueClient.repositoryInfo(ctx); err != nil {
		return fmt.Errorf("validate issue repository: %w", err)
	}
	labels, err := issueClient.labels(ctx)
	if err != nil {
		return fmt.Errorf("validate issue labels: %w", err)
	}
	available := make(map[string]struct{}, len(labels))
	for _, label := range labels {
		available[strings.ToLower(label)] = struct{}{}
	}
	for _, required := range requiredLabels {
		if _, exists := available[strings.ToLower(strings.TrimSpace(required))]; !exists {
			return errors.New("required GitHub bug report label is missing")
		}
	}
	if assetClient == nil {
		return nil
	}
	assetInfo, err := assetClient.repositoryInfo(ctx)
	if err != nil {
		return fmt.Errorf("validate asset repository: %w", err)
	}
	if assetInfo.Private {
		return errors.New("GitHub bug report asset repository must be public")
	}
	return nil
}

type githubBugReportRepositoryInfo struct {
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

func (client *GitHubBugReportIssueClient) repositoryInfo(ctx context.Context) (githubBugReportRepositoryInfo, error) {
	var info githubBugReportRepositoryInfo
	err := client.client.requestJSON(ctx, http.MethodGet, client.repository.apiPath(), nil, nil, &info, http.StatusOK)
	return info, err
}

func (client *GitHubBugReportAssetClient) repositoryInfo(ctx context.Context) (githubBugReportRepositoryInfo, error) {
	var info githubBugReportRepositoryInfo
	err := client.client.requestJSON(ctx, http.MethodGet, client.repository.apiPath(), nil, nil, &info, http.StatusOK)
	return info, err
}

func (client *GitHubBugReportIssueClient) labels(ctx context.Context) ([]string, error) {
	labels := []string{}
	for page := 1; ; page++ {
		var response []struct {
			Name string `json:"name"`
		}
		query := url.Values{"per_page": {"100"}, "page": {strconv.Itoa(page)}}
		if err := client.client.requestJSON(ctx, http.MethodGet, client.repository.apiPath("labels"), query,
			nil, &response, http.StatusOK); err != nil {
			return nil, err
		}
		for _, label := range response {
			labels = append(labels, label.Name)
		}
		if len(response) < 100 {
			return labels, nil
		}
	}
}

func (client *githubBugReportClient) requestJSON(
	ctx context.Context,
	method string,
	endpoint string,
	query url.Values,
	payload any,
	response any,
	acceptedStatuses ...int,
) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return &GitHubBugReportError{Operation: "encode request"}
		}
		body = bytes.NewReader(encoded)
	}
	requestURL := client.baseURL + endpoint
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return &GitHubBugReportError{Operation: "build request"}
	}
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", githubBugReportAPIVersion)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	httpResponse, err := client.httpClient.Do(request)
	if err != nil {
		return &GitHubBugReportError{Operation: "request"}
	}
	defer func() { _ = httpResponse.Body.Close() }()
	accepted := false
	for _, status := range acceptedStatuses {
		if httpResponse.StatusCode == status {
			accepted = true
			break
		}
	}
	if !accepted {
		return &GitHubBugReportError{Operation: "request", StatusCode: httpResponse.StatusCode}
	}
	if httpResponse.StatusCode == http.StatusNotFound {
		return &GitHubBugReportError{Operation: "request", StatusCode: http.StatusNotFound}
	}
	if response == nil {
		return nil
	}
	limited := io.LimitReader(httpResponse.Body, githubBugReportMaxResponse+1)
	encoded, err := io.ReadAll(limited)
	if err != nil || len(encoded) > githubBugReportMaxResponse {
		return &GitHubBugReportError{Operation: "read response", StatusCode: httpResponse.StatusCode}
	}
	if err := json.Unmarshal(encoded, response); err != nil {
		return &GitHubBugReportError{Operation: "decode response", StatusCode: httpResponse.StatusCode}
	}
	return nil
}

func (repository githubBugReportRepository) apiPath(parts ...string) string {
	segments := []string{"repos", repository.Owner, repository.Name}
	segments = append(segments, parts...)
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return "/" + strings.Join(segments, "/")
}

func validGitHubBugReportAssetPath(filePath string) (string, error) {
	cleaned := path.Clean(strings.ReplaceAll(filePath, "\\", "/"))
	if cleaned != filePath || !strings.HasPrefix(cleaned, "reports/") || strings.ContainsAny(cleaned, "?#\x00") {
		return "", errors.New("invalid GitHub bug report asset path")
	}
	return cleaned, nil
}

func escapeGitHubBugReportPath(filePath string) string {
	segments := strings.Split(filePath, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
