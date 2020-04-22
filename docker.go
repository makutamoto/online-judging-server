package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func runContainer(ctx context.Context, cli *client.Client, data string) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		OpenStdin:    true,
		Image:        "makutamoto/offline-judging-program",
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	hijacked, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		panic(err)
	}

	go stdcopy.StdCopy(os.Stdout, os.Stderr, hijacked.Conn)

	hijacked.Conn.Write([]byte(data + "\n\n"))

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
	cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
}

func initDocker() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, "makutamoto/offline-judging-program", types.ImagePullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, reader)
	return cli
}
