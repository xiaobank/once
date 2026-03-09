package command

import (
	"github.com/spf13/cobra"

	"github.com/basecamp/once/internal/version"
)

type updateCommand struct {
	cmd *cobra.Command
}

func newUpdateCommand() *updateCommand {
	u := &updateCommand{}
	u.cmd = &cobra.Command{
		Use:   "update",
		Short: "Update once to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.NewUpdater().UpdateBinary()
		},
	}
	return u
}
