package mocks

import (
	"context"
	"net/http"

	"github.com/google/go-github/v43/github"
	"github.com/gruntwork-io/git-xargs/auth"
)

// Mock *github.Repository slice that is returned from the mock Repositories service in test
var ownerName = "gruntwork-io"

var repoName1 = "terragrunt"
var repoName2 = "terratest"
var repoName3 = "fetch"
var repoName4 = "terraform-kubernetes-helm"

var repoURL1 = "https://github.com/gruntwork-io/terragrunt"
var repoURL2 = "https://github.com/gruntwork-io/terratest"
var repoURL3 = "https://github.com/gruntwork-io/fetch"
var repoURL4 = "https://github.com/gruntwork-io/terraform-kubernetes-helm"

var archivedFlag = true

var MockGithubRepositories = []*github.Repository{
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
	&github.Repository{
		Owner: &github.User{
			Login: &ownerName,
		},
		Name:     &repoName4,
		HTMLURL:  &repoURL4,
		Archived: &archivedFlag,
	},
}

// This mocks the PullRequest service in go-github that is used in production to call the associated GitHub endpoint
type mockGithubPullRequestService struct {
	PullRequest *github.PullRequest
	Response    *github.Response
}

func (m mockGithubPullRequestService) Create(ctx context.Context, owner, name string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return m.PullRequest, m.Response, nil
}

func (m mockGithubPullRequestService) List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	return []*github.PullRequest{m.PullRequest}, m.Response, nil
}

// This mocks the Repositories service in go-github that is used in production to call the associated GitHub endpoint
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

// ConfigureMockGithubClient returns a valid GithubClient configured for testing purposes, complete with the mocked services
func ConfigureMockGithubClient() auth.GithubClient {
	// Call the same NewClient method that is used by the actual CLI to obtain a GitHub client that calls the
	// GitHub API. In testing, however, we just implement the mock services above to satisfy the interfaces required
	// by the GithubClient. GithubClient is used uniformly between production and test code, with the only difference
	// being that in test we do not actually execute API calls to GitHub
	client := auth.NewClient(github.NewClient(nil))

	testHTMLUrl := "https://github.com/gruntwork-io/test/pull/1"

	client.Repositories = mockGithubRepositoriesService{
		Repository:   MockGithubRepositories[0],
		Repositories: MockGithubRepositories,
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

func GetMockGithubRepo() *github.Repository {
	userLogin := "gruntwork-io"
	user := &github.User{
		Login: &userLogin,
	}

	repoName := "terragrunt"
	cloneURL := "https://github.com/gruntwork-io/terragrunt"

	repo := &github.Repository{
		Owner:    user,
		Name:     &repoName,
		CloneURL: &cloneURL,
	}

	return repo
}
