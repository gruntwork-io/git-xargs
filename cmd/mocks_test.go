package main

import (
	"context"
	"net/http"

	"github.com/google/go-github/v32/github"
)

// Mock *github.Repository slice that is returned from the mock Repositories service in test
var ownerName = "gruntwork-io"

var repoName1 = "terragrunt"
var repoName2 = "terratest"
var repoName3 = "fetch"

var repoURL1 = "https://github.com/gruntwork-io/terragrunt"
var repoURL2 = "https://github.com/gruntwork-io/terratest"
var repoURL3 = "https://github.com/gruntwork-io/fetch"

var mockGithubRepositories = []*github.Repository{
	&github.Repository{
		Owner: &github.User{
			Login: &ownerName,
		},
		Name:    &repoName1,
		HTMLURL: &repoURL1,
	},
	&github.Repository{
		Owner: &github.User{
			Login: &ownerName,
		},
		Name:    &repoName2,
		HTMLURL: &repoURL2,
	},
	&github.Repository{
		Owner: &github.User{
			Login: &ownerName,
		},
		Name:    &repoName3,
		HTMLURL: &repoURL3,
	},
}

// This mocks the PullRequest service in go-github that is used in production to call the associated Github endpoint
type mockGithubPullRequestService struct {
	PullRequest *github.PullRequest
	Response    *github.Response
}

func (m mockGithubPullRequestService) Create(ctx context.Context, owner, name string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return m.PullRequest, m.Response, nil
}

// This mocks the Repositories service in go-github that is used in production to call the associated Github endpoint
type mockGithubRepositoriesService struct {
	Repository   *github.Repository
	Repositories []*github.Repository
	Response     *github.Response
}

func (m mockGithubRepositoriesService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	return m.Repository, m.Response, nil
}

func (m mockGithubRepositoriesService) ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
	return m.Repositories, m.Response, nil
}

// A convenience method to return a valid GithubClient configured for testing purposes, complete with the mocked services
func configureMockGithubClient() GithubClient {
	// Call the same NewClient method that is used by the actual CLI to obtain a Github client that calls the
	// Github API. In testing, however, we just implement the mock services above to satisfy the interfaces required
	// by the GithubClient. GithubClient is used uniformly between production and test code, with the only difference
	// being that in test we do not actually execute API calls to Github
	client := NewClient(github.NewClient(nil))

	testHTMLUrl := "https://github.com/gruntwork-io/test/pull/1"

	client.Repositories = mockGithubRepositoriesService{
		Repository:   mockGithubRepositories[0],
		Repositories: mockGithubRepositories,
		Response: &github.Response{

			Response: &http.Response{
				StatusCode: 200,
			},

			NextPage:  0,
			PrevPage:  0,
			FirstPage: 0,
			LastPage:  0,

			NextPageToken: "dontuseme",

			Rate: github.Rate{},
		},
	}
	client.PullRequests = mockGithubPullRequestService{
		PullRequest: &github.PullRequest{
			HTMLURL: &testHTMLUrl,
		},
		Response: &github.Response{},
	}

	return client
}
