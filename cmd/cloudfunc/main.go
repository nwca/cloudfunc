package main

import (
	"fmt"
	"log"
	"path"

	"github.com/nwca/cloudfunc/gcp"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "cloudfunc",
	Short: "cloud function utility for Go",
}

func init() {
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
			return gcp.Build(pkg, out)
		},
	}
	Root.AddCommand(buildCmd)

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy cloud function",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments: function name and the package")
			}
			b, _ := cmd.Flags().GetString("bucket")
			if b == "" {
				return fmt.Errorf("staging bucket not specified")
			}
			name, pkg := args[0], args[1]
			return gcp.Deploy(name, pkg, b)
		},
	}
	deployCmd.Flags().StringP("bucket", "b", "", "staging bucket to upload sources")
	Root.AddCommand(deployCmd)
}

func main() {
	if err := Root.Execute(); err != nil {
		log.Fatal(err)
	}
}
