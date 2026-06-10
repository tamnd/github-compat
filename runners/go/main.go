// Go client conformance runner. Gates GOG-1 through GOG-12.
package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	host := getenv("GITHOME_HOST", "")
	if host == "" {
		fatalf("GITHOME_HOST is required")
	}
	token := getenv("GITHOME_TOKEN", "")
	if token == "" {
		fatalf("GITHOME_TOKEN is required")
	}
	owner := getenv("GITHOME_OWNER", "test-owner")
	repo := getenv("GITHOME_REPO", "test-repo")
	insecure := os.Getenv("GITHOME_INSECURE") == "1"
	scheme := "https"
	if insecure {
		scheme = "http"
	}
	restBase := fmt.Sprintf("%s://%s/api/v3/", scheme, host)
	uploadBase := fmt.Sprintf("%s://%s/api/uploads/", scheme, host)
	gqlURL := fmt.Sprintf("%s://%s/api/graphql", scheme, host)

	appID := parseInt64(os.Getenv("GITHOME_APP_ID"))
	installID := parseInt64(os.Getenv("GITHOME_INSTALL_ID"))
	appPEM := os.Getenv("GITHOME_APP_PEM")

	tw := newTAP(12)
	defer tw.print()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	ghClient, err := github.NewEnterpriseClient(restBase, uploadBase, tc)
	if err != nil {
		fatalf("create github client: %v", err)
	}

	// GOG-1: Users.Get current user
	tw.run("GOG-1", "Users.Get current user", func() error {
		u, _, err := ghClient.Users.Get(ctx, "")
		if err != nil {
			return err
		}
		if u.GetLogin() == "" {
			return fmt.Errorf("no login")
		}
		return nil
	})

	// GOG-2: Repositories.Get
	tw.run("GOG-2", "Repositories.Get", func() error {
		r, _, err := ghClient.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return err
		}
		if r.GetName() == "" {
			return fmt.Errorf("no name")
		}
		if r.GetDefaultBranch() == "" {
			return fmt.Errorf("no default_branch")
		}
		return nil
	})

	// GOG-3: Issues.ListByRepo paginated
	tw.run("GOG-3", "Issues.ListByRepo paginated", func() error {
		opts := &github.IssueListByRepoOptions{
			State:       "all",
			ListOptions: github.ListOptions{PerPage: 5},
		}
		_, resp, err := ghClient.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return err
		}
		_ = resp.NextPage
		return nil
	})

	// GOG-4: Issues.Create
	tw.run("GOG-4", "Issues.Create", func() error {
		title := fmt.Sprintf("compat-gog4-%d", time.Now().UnixMilli())
		issue, _, err := ghClient.Issues.Create(ctx, owner, repo, &github.IssueRequest{
			Title: &title,
			Body:  github.Ptr("automated compat test"),
		})
		if err != nil {
			return err
		}
		if issue.GetNumber() <= 0 {
			return fmt.Errorf("bad issue number: %d", issue.GetNumber())
		}
		state := "closed"
		_, _, err = ghClient.Issues.Edit(ctx, owner, repo, issue.GetNumber(), &github.IssueRequest{
			State: &state,
		})
		return err
	})

	// GOG-5: RateLimit.Get
	tw.run("GOG-5", "RateLimit.Get", func() error {
		rl, _, err := ghClient.RateLimit.Get(ctx)
		if err != nil {
			return err
		}
		if rl.GetCore().Limit <= 0 {
			return fmt.Errorf("rate limit is %d", rl.GetCore().Limit)
		}
		return nil
	})

	// GOG-6: PullRequests.Get
	prNum := int(parseInt64(getenv("GITHOME_PR_NUMBER", "1")))
	tw.run("GOG-6", "PullRequests.Get", func() error {
		pr, _, err := ghClient.PullRequests.Get(ctx, owner, repo, prNum)
		if err != nil {
			return err
		}
		if pr.GetNumber() <= 0 {
			return fmt.Errorf("bad PR number")
		}
		if pr.GetState() == "" {
			return fmt.Errorf("no state")
		}
		return nil
	})

	// GOG-7: PullRequests.Merge — skip (would close fixture PR)
	tw.skip("GOG-7", "PullRequests.Merge", "skipped to preserve fixture PR")

	// GOG-8: Checks.CreateCheckRun (requires App auth)
	if appID > 0 && installID > 0 && appPEM != "" {
		tw.run("GOG-8", "Checks.CreateCheckRun", func() error {
			itr, err := ghinstallation.New(http.DefaultTransport, appID, installID, []byte(appPEM))
			if err != nil {
				return fmt.Errorf("ghinstallation: %v", err)
			}
			itr.BaseURL = restBase
			appClient, err := github.NewEnterpriseClient(restBase, uploadBase, &http.Client{Transport: itr})
			if err != nil {
				return err
			}
			name := "compat-gog8"
			status := "completed"
			conclusion := "success"
			now := github.Timestamp{Time: time.Now()}
			cr, _, err := appClient.Checks.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
				Name:        name,
				HeadSHA:     getenv("GITHOME_HEAD_SHA", "HEAD"),
				Status:      &status,
				Conclusion:  &conclusion,
				CompletedAt: &now,
			})
			if err != nil {
				return err
			}
			if cr.GetID() <= 0 {
				return fmt.Errorf("check run ID is %d", cr.GetID())
			}
			return nil
		})
	} else {
		tw.skip("GOG-8", "Checks.CreateCheckRun", "GITHOME_APP_ID/INSTALL_ID/APP_PEM not set")
	}

	// GOG-9: ETag caching via manual If-None-Match
	tw.run("GOG-9", "ETag / If-None-Match caching", func() error {
		req, _ := http.NewRequestWithContext(ctx, "GET", restBase+"repos/"+owner+"/"+repo, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp1, err := tc.Do(req)
		if err != nil {
			return err
		}
		resp1.Body.Close()
		etag := resp1.Header.Get("ETag")
		if etag == "" {
			return fmt.Errorf("no ETag on first response")
		}

		req2, _ := http.NewRequestWithContext(ctx, "GET", restBase+"repos/"+owner+"/"+repo, nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Accept", "application/vnd.github+json")
		req2.Header.Set("If-None-Match", etag)
		resp2, err := tc.Do(req2)
		if err != nil {
			return err
		}
		resp2.Body.Close()
		if resp2.StatusCode != 304 {
			return fmt.Errorf("expected 304, got %d", resp2.StatusCode)
		}
		return nil
	})

	// GOG-10: githubv4 viewer query
	tw.run("GOG-10", "githubv4 viewer query", func() error {
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
		gqlClient := githubv4.NewEnterpriseClient(gqlURL, httpClient)

		var q struct {
			Viewer struct {
				Login string
			}
		}
		if err := gqlClient.Query(ctx, &q, nil); err != nil {
			return err
		}
		if q.Viewer.Login == "" {
			return fmt.Errorf("no viewer.login")
		}
		return nil
	})

	// GOG-11: githubv4 repository query
	tw.run("GOG-11", "githubv4 repository query", func() error {
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
		gqlClient := githubv4.NewEnterpriseClient(gqlURL, httpClient)

		var q struct {
			Repository struct {
				ID githubv4.ID
			} `graphql:"repository(owner: $owner, name: $name)"`
		}
		variables := map[string]interface{}{
			"owner": githubv4.String(owner),
			"name":  githubv4.String(repo),
		}
		if err := gqlClient.Query(ctx, &q, variables); err != nil {
			return err
		}
		if q.Repository.ID == nil {
			return fmt.Errorf("no repository.id")
		}
		return nil
	})

	// GOG-12: ghinstallation token
	if appID > 0 && installID > 0 && appPEM != "" {
		tw.run("GOG-12", "ghinstallation installation token", func() error {
			itr, err := ghinstallation.New(http.DefaultTransport, appID, installID, []byte(appPEM))
			if err != nil {
				return fmt.Errorf("ghinstallation: %v", err)
			}
			itr.BaseURL = restBase
			tok, err := itr.Token(ctx)
			if err != nil {
				return err
			}
			if len(tok) < 5 {
				return fmt.Errorf("token too short: %q", tok[:min(len(tok), 10)])
			}
			return nil
		})
	} else {
		tw.skip("GOG-12", "ghinstallation installation token", "GITHOME_APP_ID/INSTALL_ID/APP_PEM not set")
	}
}

// tap writer

type tapWriter struct {
	total   int
	n       int
	entries []string
}

func newTAP(total int) *tapWriter { return &tapWriter{total: total} }

func (t *tapWriter) run(gate, desc string, fn func() error) {
	t.n++
	if err := fn(); err != nil {
		t.entries = append(t.entries, fmt.Sprintf("not ok %d - %s: %s", t.n, gate, desc))
		t.entries = append(t.entries, fmt.Sprintf("  # %v", err))
	} else {
		t.entries = append(t.entries, fmt.Sprintf("ok %d - %s: %s", t.n, gate, desc))
	}
}

func (t *tapWriter) skip(gate, desc, reason string) {
	t.n++
	t.entries = append(t.entries, fmt.Sprintf("ok %d - %s: %s # SKIP %s", t.n, gate, desc, reason))
}

func (t *tapWriter) print() {
	fmt.Printf("TAP version 14\n1..%d\n", t.total)
	for _, e := range t.entries {
		fmt.Println(e)
	}
}

// helpers

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func fatalf(f string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal: "+f+"\n", args...)
	os.Exit(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseRSAKey parses an RSA private key PEM block (used only for validation).
func parseRSAKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

var _ = parseRSAKey // used for import validation only
