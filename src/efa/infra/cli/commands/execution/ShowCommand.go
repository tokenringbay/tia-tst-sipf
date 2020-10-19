package execution

import (
	"context"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPI "efa/infra/rest/generated/client"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var limit int32
var status string
var exID string

//ShowCommand provides commands to list execution status
var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "Display the list of executions",
	RunE:  utils.TimedRunE(runExecutionList),
}

func init() {
	ShowCommand.Flags().Int32Var(&limit, "limit", 10, "Limit the number of executions to be listed. Value \"0\" will list all the executions.")
	ShowCommand.Flags().StringVar(&status, "status", "all", "Filter the executions based on the status(failed/succeeded/all)")
	ShowCommand.Flags().StringVar(&exID, "id", "", "Filter the executions based on execution id. \"limit\" and \"status\" flags are ignored when \"id\" flag is given.")
}

func runExecutionList(cmd *cobra.Command, args []string) error {

	if len(args) != 0 {
		fmt.Println("Additional arguments passed to the command.")
		cmd.Help()
		return nil
	}
	//Get base configuration
	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	if status != "all" && status != "failed" && status != "succeeded" {
		fmt.Println("Wrong value for the flag \"status\".")
		cmd.Help()
		return nil
	}

	if len(exID) == 0 {
		//Fetch Execution List from the Device
		ExecutionsResponse, _, err := api.ExecutionListApi.ExecutionList(context.Background(), limit, map[string]interface{}{"status": status})

		if err != nil {
			if utils.IsServerConnectionError(err) {
				return nil
			}
			fmt.Println(err.Error())
			return nil
		}

		//Render using Tables
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"ID", "Command", "Status", "Start Time", "End Time"})

		for _, executionLog := range ExecutionsResponse.Items {
			row := []string{executionLog.Id, executionLog.Command, executionLog.Status,
				executionLog.StartTime.Format(constants.DefaultTimeFormat),
				executionLog.EndTime.Format(constants.DefaultTimeFormat)}
			table.Append(row)
		}
		table.Render()

	} else {
		ExecutionDetails, _, err := api.ExecutionGetApi.ExecutionGet(context.Background(), exID)

		if err != nil {
			if utils.IsServerConnectionError(err) {
				return nil
			}
			fmt.Println(err.Error())
			return nil
		}

		tab := new(tabwriter.Writer)
		tab.Init(os.Stdout, 0, 8, 0, '\t', 0)

		fmt.Fprintf(tab, "ID\t:\t%s\n", ExecutionDetails.Id)
		fmt.Fprintf(tab, "COMMAND\t:\t%s\n", ExecutionDetails.Command)
		fmt.Fprintf(tab, "PARAMETERS\t:\t%s\n", ExecutionDetails.Parameters)
		fmt.Fprintf(tab, "STATUS\t:\t%s\n", ExecutionDetails.Status)
		fmt.Fprintf(tab, "START TIME\t:\t%s\n", ExecutionDetails.StartTime)
		fmt.Fprintf(tab, "END TIME\t:\t%s\n", ExecutionDetails.EndTime)
		fmt.Fprintf(tab, "LOGS\t:\n%s", ExecutionDetails.Logs)
		tab.Flush()
	}

	return nil
}
