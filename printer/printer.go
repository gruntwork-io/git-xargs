package printer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gruntwork-io/git-xargs/types"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
)

// configurePrinterStyling accepts a pointer to a table printer and sets up the styles commonly used across them
// resulting in uniform tabular output to STDOUT following each run of the CLI
func configurePrinterStyling(printer *tableprinter.Printer) {
	printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
	printer.CenterSeparator = "│"
	printer.ColumnSeparator = "│"
	printer.RowSeparator = "─"
	printer.HeaderBgColor = tablewriter.BgBlackColor
	printer.HeaderFgColor = tablewriter.FgGreenColor
}

func PrintRepoReport(allEvents []types.AnnotatedEvent, runReport *types.RunReport) {
	fmt.Print("\n\n")
	fmt.Println("*****************************************************************")
	fmt.Printf("  GIT-XARGS RUN SUMMARY @ %v\n", time.Now().UTC())
	fmt.Printf("  Runtime in seconds: %v\n", runReport.RuntimeSeconds)
	fmt.Println("*****************************************************************")

	// If there were any allowed repos provided via file, print out the list of them
	fileProvidedReposPrinter := tableprinter.New(os.Stdout)
	configurePrinterStyling(fileProvidedReposPrinter)

	fmt.Print("\n\n")

	fmt.Println("COMMAND SUPPLIED")
	fmt.Println()
	fmt.Println(runReport.Command)
	fmt.Println()
	fmt.Println("REPO SELECTION METHOD USED FOR THIS RUN - (see README.md for more information)")
	fmt.Println()
	fmt.Println(runReport.SelectionMode)

	// If the user selected repos via a flat file, print a table showing which repos they were
	if len(runReport.FileProvidedRepos) > 0 {
		fmt.Println(" REPOS SUPPLIED VIA --repos FILE FLAG")
		fileProvidedReposPrinter.Print(runReport.FileProvidedRepos)
	}
	// For each event type, print a summary of the repos in that category
	for _, ae := range allEvents {

		var reducedRepos []types.ReducedRepo

		printer := tableprinter.New(os.Stdout)
		configurePrinterStyling(printer)

		for _, repo := range runReport.Repos[ae.Event] {
			rr := types.ReducedRepo{
				Name: repo.GetName(),
				URL:  repo.GetHTMLURL(),
			}
			reducedRepos = append(reducedRepos, rr)
		}

		if len(reducedRepos) > 0 {
			fmt.Println()
			fmt.Printf(" %s\n", strings.ToUpper(ae.Description))
			printer.Print(reducedRepos)
			fmt.Println()
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

	if len(pullRequests) > 0 {
		fmt.Println()
		fmt.Println("*****************************************************")
		fmt.Println("  PULL REQUESTS OPENED")
		fmt.Println("*****************************************************")
		pullRequestPrinter := tableprinter.New(os.Stdout)
		configurePrinterStyling(pullRequestPrinter)
		pullRequestPrinter.Print(pullRequests)
		fmt.Println()

	}
}
