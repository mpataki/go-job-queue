package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	jobqueuev1 "github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1"
	"github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1/jobqueuev1connect"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "job",
		Short: "Job queue CLI",
	}

	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "Job Server address")

	rootCmd.AddCommand(newSubmitCommand())
	rootCmd.Execute()
}

func newSubmitCommand() *cobra.Command {
	submitCmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a job",
		Run: func(cmd *cobra.Command, args []string) {
			serverAddr, _ := cmd.Root().PersistentFlags().GetString("server")

			jobType, _ := cmd.Flags().GetString("type")
			payload, _ := cmd.Flags().GetString("payload")
			at, _ := cmd.Flags().GetInt64("at")

			client := jobqueuev1connect.NewJobServiceClient(http.DefaultClient, serverAddr)

			ctx := context.Background()

			resp, err := client.EnqueueJob(ctx, connect.NewRequest(&jobqueuev1.EnqueueJobRequest{
				Type:            jobType,
				Payload:         []byte(payload),
				ExecutionTimeMs: &at,
			}))
			if err != nil {
				log.Fatalf("Error sending enqueue request to service: %v", err)
			}

			data, _ := json.MarshalIndent(resp.Msg, "", "  ")
			fmt.Println(string(data))
		},
	}

	submitCmd.Flags().String("type", "", "Job type")
	submitCmd.Flags().String("payload", "", "Job payload")
	submitCmd.Flags().Int64("at", time.Now().UnixMilli(), "Execution time")
	submitCmd.MarkFlagRequired("type")
	submitCmd.MarkFlagRequired("payload")

	return submitCmd
}
