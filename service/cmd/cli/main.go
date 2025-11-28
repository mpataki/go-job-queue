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
	rootCmd.AddCommand(newGetJobCommand())
	rootCmd.AddCommand(newCancelJobCommand())
	rootCmd.Execute()
}

func newSubmitCommand() *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.Flags().String("type", "", "Job type")
	cmd.Flags().String("payload", "", "Job payload")
	cmd.Flags().Int64("at", time.Now().UnixMilli(), "Execution time")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("payload")

	return cmd
}

func newGetJobCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a job",
		Run: func(cmd *cobra.Command, args []string) {
			serverAddr, _ := cmd.Root().PersistentFlags().GetString("server")

			id, _ := cmd.Flags().GetString("id")

			client := jobqueuev1connect.NewJobServiceClient(http.DefaultClient, serverAddr)

			ctx := context.Background()

			resp, err := client.GetJob(ctx, connect.NewRequest(&jobqueuev1.GetJobRequest{
				Id: id,
			}))
			if err != nil {
				log.Fatalf("Error fetching job from service: %v", err)
			}

			data, _ := json.MarshalIndent(resp.Msg, "", "  ")
			fmt.Println(string(data))
		},
	}

	cmd.Flags().String("id", "", "Job ID")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newCancelJobCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a job",
		Run: func(cmd *cobra.Command, args []string) {
			serverAddr, _ := cmd.Root().PersistentFlags().GetString("server")

			id, _ := cmd.Flags().GetString("id")

			client := jobqueuev1connect.NewJobServiceClient(http.DefaultClient, serverAddr)

			ctx := context.Background()

			resp, err := client.CancelJob(ctx, connect.NewRequest(&jobqueuev1.CancelJobRequest{
				Id: id,
			}))
			if err != nil {
				log.Fatalf("Error cancelling a job: %v", err)
			}

			data, _ := json.MarshalIndent(resp.Msg, "", "  ")
			fmt.Println(string(data))
		},
	}

	cmd.Flags().String("id", "", "Job ID")
	cmd.MarkFlagRequired("id")

	return cmd
}
