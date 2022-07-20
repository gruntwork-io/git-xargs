package printer

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/git-xargs/types"
	"github.com/pterm/pterm"
)

func PrintRepoReport(allEvents []types.AnnotatedEvent, runReport *types.RunReport) {
	renderSection(fmt.Sprintf("Git-xargs run summary @ %s", time.Now().UTC()))

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Runtime in seconds: %d", runReport.RuntimeSeconds)},
		{Level: 0, Text: fmt.Sprintf("Command supplied: %s", runReport.Command)},
		{Level: 0, Text: fmt.Sprintf("Repo selection method: %s", runReport.SelectionMode)},
	}).Render()

	if len(runReport.FileProvidedRepos) > 0 {
		renderSection("Repos supplied via --repos file flag")
		data := make([][]string, len(runReport.FileProvidedRepos))
		for idx, fileProvidedRepo := range runReport.FileProvidedRepos {
			data[idx] = []string{fmt.Sprintf("%s/%s", fileProvidedRepo.Organization, fileProvidedRepo.Name)}
		}
		renderTableWithHeader([]string{"Repo name"}, data)
	}

	// For each event type, print a summary table of the repos in that category
	for _, ae := range allEvents {

		var reducedRepos []types.ReducedRepo

		for _, repo := range runReport.Repos[ae.Event] {
			rr := types.ReducedRepo{
				Name: repo.GetName(),
				URL:  repo.GetHTMLURL(),
			}
			reducedRepos = append(reducedRepos, rr)
		}

		if len(reducedRepos) > 0 {

			renderSection(ae.Description)
			data := make([][]string, len(reducedRepos))
			for idx, repo := range reducedRepos {
				data[idx] = []string{repo.Name, repo.URL}
			}

			renderTableWithHeader([]string{"Repo name", "Repo URL"}, data)
		}
	}

	var pullRequests []types.PullRequest

	for repoName, prURL := range runReport.PullRequests {
		pr := types.PullRequest{
			Repo: repoName,
			URL:  prURL,
		}
		pullRequests = append(pullRequests, pr)
	}

	var draftPullRequests []types.PullRequest

	for repoName, prURL := range runReport.DraftPullRequests {
		pr := types.PullRequest{
			Repo: repoName,
			URL:  prURL,
		}
		draftPullRequests = append(draftPullRequests, pr)
	}

	if len(pullRequests) > 0 {
		renderSection("Pull requests opened")

		data := make([][]string, len(pullRequests))
		for idx, pullRequest := range pullRequests {
			data[idx] = []string{pullRequest.Repo, pullRequest.URL}
		}

		renderTableWithHeader([]string{"Repo name", "Pull request URL"}, data)
	}

	if len(draftPullRequests) > 0 {
		renderSection("Draft Pull requests opened")

		data := make([][]string, len(draftPullRequests))
		for idx, draftPullRequest := range draftPullRequests {
			data[idx] = []string{draftPullRequest.Repo, draftPullRequest.URL}
		}

		renderTableWithHeader([]string{"Repo name", "Draft Pull request URL"}, data)
	}
}

func renderSection(sectionTitle string) {
	pterm.DefaultSection.Style = pterm.NewStyle(pterm.FgLightCyan)
	pterm.DefaultSection.WithLevel(0).Println(sectionTitle)
}

func renderTableWithHeader(headers []string, data [][]string) {
	tableData := pterm.TableData{
		headers,
	}
	for idx := range data {
		tableData = append(tableData, data[idx])
	}
	pterm.DefaultTable.
		WithHasHeader().
		WithBoxed(true).
		WithRowSeparator("-").
		WithData(tableData).
		Render()
}
