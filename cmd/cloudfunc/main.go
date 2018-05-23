package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/nwca/cloudfunc/gcp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var Root = &cobra.Command{
	Use:   "cloudfunc",
	Short: "cloud function utility for Go",
}

func init() {
	const (
		projectFlag   = "project"
		appConfigFlag = "app-config"
	)
	Root.PersistentFlags().StringP(projectFlag, "p", "", "project to use")

	getDeployParams := func(cmd *cobra.Command) (cli *gcp.Client, env map[string]string, _ error) {
		proj, _ := cmd.Flags().GetString(projectFlag)
		if proj == "" {
			return nil, nil, fmt.Errorf("project not specified")
		}
		if c, _ := cmd.Flags().GetString(appConfigFlag); c != "" {
			data, err := ioutil.ReadFile(c)
			if err != nil {
				return nil, nil, err
			}

			var conf struct {
				Env map[string]string `yaml:"env_variables"`
			}

			if err := yaml.Unmarshal(data, &conf); err != nil {
				return nil, nil, err
			}
			env = conf.Env
		}
		cli, err := gcp.NewClient(proj)
		if err != nil {
			return nil, nil, err
		}
		return cli, env, nil
	}

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy cloud function",
	}
	deployCmd.PersistentFlags().StringP(appConfigFlag, "c", "", "app config to use")
	Root.AddCommand(deployCmd)

	deployTrigger := func(cmd *cobra.Command, name string, tr gcp.Trigger) error {
		ctx := context.Background()
		cli, env, err := getDeployParams(cmd)
		if err != nil {
			return err
		}
		defer cli.Close()

		file, err := gcp.BuildTmp(tr, env)
		if err != nil {
			return err
		}
		defer file.Close()

		log.Println("deploying function", name)
		return cli.Deploy(ctx, name, tr, file)
	}

	deployZip := &cobra.Command{
		Use:   "zip",
		Short: "deploy zip file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and archive name")
			}
			ctx := context.Background()
			name, pkg := args[0], args[1]

			f, err := os.Open(pkg)
			if err != nil {
				return err
			}
			defer f.Close()

			cli, _, err := getDeployParams(cmd)
			if err != nil {
				return err
			}
			defer cli.Close()

			return cli.Deploy(ctx, name, gcp.HTTPTrigger{}, f)
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
			name, pkg := args[0], args[1]
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}

			return deployTrigger(cmd, name, gcp.HTTPTrigger{Target: t})
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
			topic, _ := cmd.Flags().GetString("topic")
			if topic == "" {
				return fmt.Errorf("topic not specified")
			}
			name, pkg := args[0], args[1]
			t, err := gcp.ParseTarget(pkg)
			if err != nil {
				return err
			}

			return deployTrigger(cmd, name, gcp.TopicTrigger{Target: t, Topic: topic})
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
			bucket, _ := cmd.Flags().GetString("bucket")
			if bucket == "" {
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
			return deployTrigger(cmd, name, gcp.StorageTrigger{
				Target: t, Bucket: bucket,
				//Event: gcp.StorageEvent(event),
			})
		},
	}
	deployStorage.Flags().StringP("bucket", "b", "", "bucket name")
	//deployStorage.Flags().StringP("event", "e", "", "event type")
	deployCmd.AddCommand(deployStorage)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list deployed functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected project name")
			}
			proj := args[0]
			cli, err := gcp.NewClient(proj)
			if err != nil {
				return err
			}
			list, err := cli.ListFuncs()
			if err != nil {
				return err
			}
			for _, f := range list {
				fmt.Println(f.Name)
			}
			return nil
		},
	}
	Root.AddCommand(listCmd)
}

func main() {
	if err := Root.Execute(); err != nil {
		log.Fatal(err)
	}
}
