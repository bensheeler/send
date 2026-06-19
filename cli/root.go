package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/bensheeler/send/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
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

			result, err := app.ScanRequestFile(app.ScanRequestFileInput{
				CWD:      cwd,
				Selector: args[0],
			})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(stdout, result.Path)
			return err
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	return cmd
}
