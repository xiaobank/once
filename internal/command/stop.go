package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/basecamp/once/internal/docker"
)

type stopCommand struct {
	cmd *cobra.Command
}

func newStopCommand() *stopCommand {
	s := &stopCommand{}
	s.cmd = &cobra.Command{
		Use:   "stop <app>",
		Short: "Stop an application",
		Args:  cobra.ExactArgs(1),
		RunE:  WithNamespace(s.run),
	}
	return s
}

// Private

func (s *stopCommand) run(ctx context.Context, ns *docker.Namespace, cmd *cobra.Command, args []string) error {
	appName := args[0]

	err := withApplication(ns, appName, "stopping", func(app *docker.Application) error {
		return app.Stop(ctx)
	})
	if err != nil {
		return err
	}

	fmt.Printf("Stopped %s\n", appName)
	return nil
}
