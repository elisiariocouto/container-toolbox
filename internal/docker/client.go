// Package docker talks to the local Docker daemon.
//
// It is split in two: client.go uses the Docker Go SDK over the local socket to
// read live container status, and compose.go shells out to the `docker compose`
// CLI for lifecycle actions and log streaming.
package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Compose label keys stamped on every container created by Compose.
const (
	labelProject    = "com.docker.compose.project"
	labelWorkingDir = "com.docker.compose.project.working_dir"
)

// Client wraps the Docker SDK client.
type Client struct {
	api *client.Client
}

// ProjectStatus summarizes the containers belonging to one Compose project.
type ProjectStatus struct {
	WorkingDir string
	Running    int
	Total      int
}

// NewClient creates a Docker SDK client using the standard environment
// (defaulting to the local unix socket) and negotiates the API version so it
// works against older or newer daemons.
func NewClient() (*Client, error) {
	api, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

// Close releases the underlying client.
func (c *Client) Close() error {
	if c == nil || c.api == nil {
		return nil
	}
	return c.api.Close()
}

// ListContainersByProject returns the status of every Compose project that has
// containers (running or not), keyed by the project name. It also records each
// project's working_dir label as a secondary join key.
func (c *Client) ListContainersByProject(ctx context.Context) (map[string]ProjectStatus, error) {
	args := filters.NewArgs(filters.Arg("label", labelProject))
	list, err := c.api.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, err
	}

	out := make(map[string]ProjectStatus)
	for _, ct := range list {
		proj := ct.Labels[labelProject]
		if proj == "" {
			continue
		}
		ps := out[proj]
		ps.Total++
		if ct.State == "running" {
			ps.Running++
		}
		if ps.WorkingDir == "" {
			ps.WorkingDir = ct.Labels[labelWorkingDir]
		}
		out[proj] = ps
	}
	return out, nil
}
