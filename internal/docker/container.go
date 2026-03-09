package docker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const ContainerLogMaxSize = "10m"

func ContainerLogConfig() container.LogConfig {
	return container.LogConfig{
		Type: "json-file",
		Config: map[string]string{
			"max-size": ContainerLogMaxSize,
			"max-file": "1",
		},
	}
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func execInContainer(ctx context.Context, c *client.Client, containerName string, cmd []string) (ExecResult, error) {
	execResp, err := c.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return ExecResult{}, fmt.Errorf("creating exec: %w", err)
	}

	resp, err := c.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return ExecResult{}, fmt.Errorf("attaching exec: %w", err)
	}
	defer resp.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, resp.Reader); err != nil {
		return ExecResult{}, fmt.Errorf("reading exec output: %w", err)
	}

	inspect, err := c.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return ExecResult{}, fmt.Errorf("inspecting exec: %w", err)
	}

	return ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: inspect.ExitCode,
	}, nil
}
