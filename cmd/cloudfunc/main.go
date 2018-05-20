package main

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/nwca/cloudfunc/gcp"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "cloudfunc",
	Short: "cloud function utility for Go",
}

func init() {
	const stagingBucketFlag = "staging"
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "builds cloud function from a Go package",
		RunE: func(cmd *cobra.Command, args []string) error {
			if n := len(args); n != 1 && n != 2 {
				return fmt.Errorf("expected one or two arguments")
			}
			pkg := args[0]
			out := path.Base(pkg) + ".zip"
			if len(args) > 1 {
				out = args[1]
			}
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}
			return gcp.Build(gcp.HTTPTrigger{Target: t}, out)
		},
	}
	Root.AddCommand(buildCmd)

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy cloud function",
	}
	deployCmd.PersistentFlags().StringP(stagingBucketFlag, "s", "", "staging bucket to upload sources")
	Root.AddCommand(deployCmd)

	deployZip := &cobra.Command{
		Use:   "zip",
		Short: "deploy zip file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and archive name")
			}
			b, _ := cmd.Flags().GetString(stagingBucketFlag)
			if b == "" {
				return fmt.Errorf("staging bucket not specified")
			}
			name, pkg := args[0], args[1]
			return gcp.DeployZIP(name, pkg, b)
		},
	}
	deployCmd.AddCommand(deployZip)

	deployHttp := &cobra.Command{
		Use:   "http",
		Short: "deploy http trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and package name")
			}
			b, _ := cmd.Flags().GetString(stagingBucketFlag)
			if b == "" {
				return fmt.Errorf("staging bucket not specified")
			}
			name, pkg := args[0], args[1]
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}
			return gcp.Deploy(name, gcp.HTTPTrigger{Target: t}, b)
		},
	}
	deployCmd.AddCommand(deployHttp)

	deployPubSub := &cobra.Command{
		Use:   "pubsub",
		Short: "deploy pubsub trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and package name")
			}
			b, _ := cmd.Flags().GetString(stagingBucketFlag)
			if b == "" {
				return fmt.Errorf("staging bucket not specified")
			}
			topic, _ := cmd.Flags().GetString("topic")
			if b == "" {
				return fmt.Errorf("topic not specified")
			}
			name, pkg := args[0], args[1]
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}
			return gcp.Deploy(name, gcp.TopicTrigger{Target: t, Topic: topic}, b)
		},
	}
	deployPubSub.Flags().StringP("topic", "t", "", "topic id")
	deployCmd.AddCommand(deployPubSub)

	deployStorage := &cobra.Command{
		Use:   "storage",
		Short: "deploy storage trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and package name")
			}
			b, _ := cmd.Flags().GetString(stagingBucketFlag)
			if b == "" {
				return fmt.Errorf("staging bucket not specified")
			}
			bucket, _ := cmd.Flags().GetString("bucket")
			if b == "" {
				return fmt.Errorf("topic not specified")
			}
			event, _ := cmd.Flags().GetString("event")
			if event != "" && !strings.HasPrefix(event, gcp.StorageEventPref) {
				event = gcp.StorageEventPref + event
			}
			name, pkg := args[0], args[1]
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}
			return gcp.Deploy(name, gcp.StorageTrigger{Target: t, Bucket: bucket, Event: gcp.StorageEvent(event)}, b)
		},
	}
	deployStorage.Flags().StringP("bucket", "b", "", "bucket name")
	deployStorage.Flags().StringP("event", "e", "", "event type")
	deployCmd.AddCommand(deployStorage)
}

func main() {
	if err := Root.Execute(); err != nil {
		log.Fatal(err)
	}
}
