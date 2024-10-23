package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v43/github"
	"github.com/gruntwork-io/git-xargs/types"
	"github.com/gruntwork-io/go-commons/errors"
	"golang.org/x/oauth2"
)

// The go-github package satisfies this PullRequest service's interface in production
type githubPullRequestService interface {
	Create(ctx context.Context, owner string, name string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	RequestReviewers(ctx context.Context, owner, repo string, number int, reviewers github.ReviewersRequest) (*github.PullRequest, *github.Response, error)
}

// The go-github package satisfies this Repositories service's interface in production
type githubRepositoriesService interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
}

type githubIssuesService interface {
	AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, *github.Response, error)
}

// GithubClient is the data structure that is common between production code and test code. In production code,
// go-github satisfies the PullRequests and Repositories service interfaces, whereas in test the concrete
// implementations for these same services are mocks that return a static slice of pointers to GitHub repositories,
// or a single pointer to a GitHub repository, as appropriate. This allows us to test the workflow of git-xargs
// without actually making API calls to GitHub when running tests
type GithubClient struct {
	PullRequests githubPullRequestService
	Repositories githubRepositoriesService
	Issues       githubIssuesService
}

func NewClient(client *github.Client) GithubClient {
	return GithubClient{
		PullRequests: client.PullRequests,
		Repositories: client.Repositories,
		Issues:       client.Issues,
	}
}

// ConfigureGithubClient creates a GitHub API client using the user-supplied GITHUB_OAUTH_TOKEN and returns the configured GitHub client
func ConfigureGithubClient() GithubClient {
	// Ensure user provided a GITHUB_OAUTH_TOKEN
	GithubOauthToken := os.Getenv("GITHUB_OAUTH_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubOauthToken},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	var githubClient *github.Client

	if os.Getenv("GITHUB_HOSTNAME") != "" {
		GithubHostname := os.Getenv("GITHUB_HOSTNAME")
		baseUrl := fmt.Sprintf("https://%s/", GithubHostname)

		githubClient, _ = github.NewEnterpriseClient(baseUrl, baseUrl, tc)

	} else {
		githubClient = github.NewClient(tc)
	}

	// Wrap the go-github client in a GithubClient struct, which is common between production and test code
	client := NewClient(githubClient)

	return client
}

// EnsureGithubOauthTokenSet is a sanity check that a value is exported for GITHUB_OAUTH_TOKEN
func EnsureGithubOauthTokenSet() error {
	if os.Getenv("GITHUB_OAUTH_TOKEN") == "" {
		return errors.WithStackTrace(types.NoGithubOauthTokenProvidedErr{})
	}
	return nil
}
