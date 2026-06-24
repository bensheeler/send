package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/bensheeler/send/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:          "send <request-file>",
		Short:        "Find and send HTTP requests",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := app.LoadRequest(app.LoadRequestInput{
				CWD:      cwd,
				Selector: args[0],
			})
			if err != nil {
				return err
			}

			if debug {
				_, err = fmt.Fprintln(stdout, result.Path)
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(stdout, "%s %s\n", result.Method, result.URL)
			return err
		},
	}
	cmd.Flags().BoolVar(&debug, "debug", false, "print debug information")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	return cmd
}
