package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/websocket"
)

type statusType struct {
	WholeResult int    `json:"whole_result"`
	Result      int    `json:"result"`
	Time        int64  `json:"time"`
	CurrentCase int    `json:"current_case"`
	WholeCase   int    `json:"whole_case"`
	Description string `json:"description"`
}

var cli *client.Client

func runContainer(data string, conn *websocket.Conn) {
	ctx := context.Background()
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

	hijacked.Conn.Write([]byte(data + "\n\n"))

	stdReader, stdWriter := io.Pipe()
	go stdcopy.StdCopy(stdWriter, os.Stderr, hijacked.Conn)
	scanner := bufio.NewScanner(stdReader)

	for scanner.Scan() && len(scanner.Bytes()) != 0 {
		if err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}

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

func initDocker() {
	_cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli = _cli
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, "makutamoto/offline-judging-program", types.ImagePullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, reader)
}
