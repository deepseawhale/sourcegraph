package zoekt

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Reindex forces indexserver to reindex the repo immediately.
func Reindex(ctx context.Context, name api.RepoName, id api.RepoID) error {
	// Find the Zoekt webserver hosting the index of the repo.
	ep, err := search.Indexers().Map.Get(string(name))
	if err != nil {
		return err
	}

	// We add http:// on a best-effort basis, because it is not guaranteed that
	// ep is a valid URL.
	if !strings.HasPrefix(ep, "http://") {
		ep = "http://" + ep
	}
	u, err := url.Parse(ep)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Add("repo", strconv.Itoa(int(id)))

	// http://<host:port>/indexerver/?headless
	u = u.ResolveReference(&url.URL{Path: "/indexserver/", RawQuery: "headless"})

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(b))
	}

	return nil
}