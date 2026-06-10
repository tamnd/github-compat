package fixture

import (
	"fmt"
	"os"
	"strconv"
)

// Fixture holds all connection parameters and pre-seeded object IDs
// that every runner reads from environment variables.
type Fixture struct {
	Host          string
	SSHHost       string
	HTTPSCloneURL string
	SSHCloneURL   string
	Owner         string
	Repo          string
	PAT           string

	AppID     int64
	InstallID int64
	AppPEM    string

	IssueNumber   int64
	PRNumber      int64
	HeadSHA       string
	ReleaseName   string
	WebhookURL    string
	WebhookSecret string

	RESTBase string
	GQLBase  string
}

// FromEnv reads the fixture from well-known environment variables.
func FromEnv() (*Fixture, error) {
	host := os.Getenv("GITHOME_HOST")
	if host == "" {
		return nil, fmt.Errorf("GITHOME_HOST is required")
	}
	token := os.Getenv("GITHOME_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHOME_TOKEN is required")
	}

	scheme := "https"
	if os.Getenv("GITHOME_INSECURE") == "1" {
		scheme = "http"
	}

	f := &Fixture{
		Host:          host,
		SSHHost:       env("GITHOME_SSH_HOST", host+":22"),
		HTTPSCloneURL: env("GITHOME_HTTPS_CLONE_URL", fmt.Sprintf("%s://%s/%s/%s.git", scheme, host, env("GITHOME_OWNER", "test-owner"), env("GITHOME_REPO", "test-repo"))),
		SSHCloneURL:   env("GITHOME_SSH_CLONE_URL", fmt.Sprintf("git@%s:%s/%s.git", host, env("GITHOME_OWNER", "test-owner"), env("GITHOME_REPO", "test-repo"))),
		Owner:         env("GITHOME_OWNER", "test-owner"),
		Repo:          env("GITHOME_REPO", "test-repo"),
		PAT:           token,
		AppID:         envInt64("GITHOME_APP_ID", 0),
		InstallID:     envInt64("GITHOME_INSTALL_ID", 0),
		AppPEM:        os.Getenv("GITHOME_APP_PEM"),
		IssueNumber:   envInt64("GITHOME_ISSUE_NUMBER", 1),
		PRNumber:      envInt64("GITHOME_PR_NUMBER", 1),
		HeadSHA:       os.Getenv("GITHOME_HEAD_SHA"),
		ReleaseName:   env("GITHOME_RELEASE_NAME", "v0.0.1"),
		WebhookURL:    os.Getenv("GITHOME_WEBHOOK_URL"),
		WebhookSecret: os.Getenv("GITHOME_WEBHOOK_SECRET"),
		RESTBase:      fmt.Sprintf("%s://%s/api/v3", scheme, host),
		GQLBase:       fmt.Sprintf("%s://%s/api/graphql", scheme, host),
	}
	return f, nil
}

// Env returns a slice of KEY=VALUE strings suitable for os/exec.Cmd.Env.
func (f *Fixture) Env() []string {
	scheme := "https"
	if f.HTTPSCloneURL != "" && len(f.HTTPSCloneURL) >= 7 && f.HTTPSCloneURL[:7] == "http://" {
		scheme = "http"
	}
	return []string{
		"GITHOME_HOST=" + f.Host,
		"GITHOME_SSH_HOST=" + f.SSHHost,
		"GITHOME_HTTPS_CLONE_URL=" + f.HTTPSCloneURL,
		"GITHOME_SSH_CLONE_URL=" + f.SSHCloneURL,
		"GITHOME_OWNER=" + f.Owner,
		"GITHOME_REPO=" + f.Repo,
		"GITHOME_TOKEN=" + f.PAT,
		"GITHOME_APP_ID=" + strconv.FormatInt(f.AppID, 10),
		"GITHOME_INSTALL_ID=" + strconv.FormatInt(f.InstallID, 10),
		"GITHOME_APP_PEM=" + f.AppPEM,
		"GITHOME_ISSUE_NUMBER=" + strconv.FormatInt(f.IssueNumber, 10),
		"GITHOME_PR_NUMBER=" + strconv.FormatInt(f.PRNumber, 10),
		"GITHOME_HEAD_SHA=" + f.HeadSHA,
		"GITHOME_RELEASE_NAME=" + f.ReleaseName,
		"GITHOME_WEBHOOK_URL=" + f.WebhookURL,
		"GITHOME_WEBHOOK_SECRET=" + f.WebhookSecret,
		"GITHOME_REST_BASE=" + f.RESTBase,
		"GITHOME_GQL_ENDPOINT=" + f.GQLBase,
		// Pass through for tools that use standard env vars
		"GITHUB_TOKEN=" + f.PAT,
		"GH_TOKEN=" + f.PAT,
		"GH_HOST=" + f.Host,
		"GH_ENTERPRISE_TOKEN=" + f.PAT,
		"GITHUB_SERVER_URL=" + scheme + "://" + f.Host,
		"GITHUB_API_URL=" + f.RESTBase,
		"GITHUB_GRAPHQL_URL=" + f.GQLBase,
		// Inherit PATH so runners can find tools
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}
