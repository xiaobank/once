package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/basecamp/once/internal/docker"
)

type startCommand struct {
	cmd *cobra.Command
}

func newStartCommand() *startCommand {
	s := &startCommand{}
	s.cmd = &cobra.Command{
		Use:   "start <app>",
		Short: "Start an application",
		Args:  cobra.ExactArgs(1),
		RunE:  WithNamespace(s.run),
	}
	return s
}

// Private

func (s *startCommand) run(ctx context.Context, ns *docker.Namespace, cmd *cobra.Command, args []string) error {
	appName := args[0]

	err := withApplication(ns, appName, "starting", func(app *docker.Application) error {
		return app.Start(ctx)
	})
	if err != nil {
		return err
	}

	fmt.Printf("Started %s\n", appName)
	return nil
}
