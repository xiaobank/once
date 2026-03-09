package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/basecamp/once/internal/docker"
)

type listCommand struct {
	cmd *cobra.Command
}

func newListCommand() *listCommand {
	l := &listCommand{}
	l.cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed applications",
		RunE:    WithNamespace(l.run),
	}
	return l
}

// Private

func (l *listCommand) run(_ context.Context, ns *docker.Namespace, cmd *cobra.Command, args []string) error {
	for _, app := range ns.Applications() {
		fmt.Println(app.Settings.Name)
	}

	return nil
}
