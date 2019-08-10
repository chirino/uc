package findutil

import (
	"context"
	"github.com/google/go-github/v27/github"
	"io"
	"net/http"
	"net/url"
)

type GithubUtils struct {
	Context                   context.Context
	GithubUser                string
	GithubPersonalAccessToken string
}

func (o *GithubUtils) ListGithubReleases(owner string, repo string, action func(*github.RepositoryRelease) error) error {
	httpClient := &http.Client{}
	if o.GithubUser != "" {
		tp := github.BasicAuthTransport{
			Username: o.GithubUser,
			Password: o.GithubPersonalAccessToken,
		}
		httpClient = tp.Client()
	}
	client := github.NewClient(httpClient)
	options := &github.ListOptions{PerPage: 20}
	for {
		releases, resp, err := client.Repositories.ListReleases(o.Context, owner, repo, options)
		if err != nil {
			return err
		}
		for _, release := range releases {
			err := action(release)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return nil
}

func (o *GithubUtils) AuthURL(downloadURL string) string {
	if o.GithubUser == "" {
		return downloadURL
	}
	parse, err := url.Parse(downloadURL)
	if err != nil {
		panic(err)
	}
	parse.User = url.UserPassword(o.GithubUser, o.GithubPersonalAccessToken)
	return parse.String()
}
