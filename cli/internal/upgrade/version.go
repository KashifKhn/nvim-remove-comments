package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	GitHubAPIURL   = "https://api.github.com/repos/%s/%s/releases/latest"
	RequestTimeout = 30 * time.Second
	UserAgent      = "rmc"
	MaxRetries     = 3
	RetryDelay     = 2 * time.Second
)

type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
}

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Original   string
}

func isRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

func GetLatestVersion() (string, error) {
	url := fmt.Sprintf(GitHubAPIURL, RepoOwner, RepoName)
	client := &http.Client{Timeout: RequestTimeout}

	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(RetryDelay)
		}

		release, err := fetchLatestRelease(client, url)
		if err == nil {
			return release.TagName, nil
		}

		lastErr = err
		if !isRetryableError(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", MaxRetries, lastErr)
}

func fetchLatestRelease(client *http.Client, url string) (*GitHubRelease, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, &retryableError{err: fmt.Errorf("failed to fetch latest release: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s/%s", RepoOwner, RepoName)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
	}

	if isRetryableStatusCode(resp.StatusCode) {
		return nil, &retryableError{err: fmt.Errorf("GitHub API returned status %d", resp.StatusCode)}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("no version tag found in release")
	}

	return &release, nil
}

type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func isRetryableError(err error) bool {
	var re *retryableError
	return err != nil && (errors.As(err, &re))
}

func ParseVersion(v string) (*Version, error) {
	original := v
	v = strings.TrimPrefix(v, "v")

	var prerelease string
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		prerelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")

	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid version format: %s", original)
	}

	version := &Version{
		Original:   original,
		Prerelease: prerelease,
	}

	if len(parts) >= 1 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid major version: %s", parts[0])
		}
		version.Major = major
	}

	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
		version.Minor = minor
	}

	if len(parts) >= 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
		version.Patch = patch
	}

	return version, nil
}

func CompareVersions(current, latest string) (int, error) {
	if current == "dev" || current == "" {
		return -1, nil
	}

	currentV, err := ParseVersion(current)
	if err != nil {
		return 0, fmt.Errorf("failed to parse current version: %w", err)
	}

	latestV, err := ParseVersion(latest)
	if err != nil {
		return 0, fmt.Errorf("failed to parse latest version: %w", err)
	}

	if currentV.Major != latestV.Major {
		if currentV.Major < latestV.Major {
			return -1, nil
		}
		return 1, nil
	}

	if currentV.Minor != latestV.Minor {
		if currentV.Minor < latestV.Minor {
			return -1, nil
		}
		return 1, nil
	}

	if currentV.Patch != latestV.Patch {
		if currentV.Patch < latestV.Patch {
			return -1, nil
		}
		return 1, nil
	}

	if currentV.Prerelease != "" && latestV.Prerelease == "" {
		return -1, nil
	}
	if currentV.Prerelease == "" && latestV.Prerelease != "" {
		return 1, nil
	}

	return 0, nil
}

func IsNewerAvailable(current, latest string) (bool, error) {
	cmp, err := CompareVersions(current, latest)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

func (v *Version) String() string {
	if v.Prerelease != "" {
		return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func NormalizeVersion(v string) string {
	if v == "" || v == "dev" {
		return v
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
