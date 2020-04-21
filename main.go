package main

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// reader, err := cli.ImagePull(ctx, "makutamoto/offline-judging-program", types.ImagePullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// io.Copy(os.Stdout, reader)

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

	hijacked.Conn.Write([]byte(`{
		"language": "golang",
		"code": "package main\n\nimport \"fmt\"\n\nfunc fizzBuzz(number int) {\n\tif number%15 == 0 {\n\t} else if number%5 == 0 {\t\t} else if number%3 == 0 {\n\t\tfmt.Println(\"Fizz\")\n\t} else {\n\t\tfmt.Println(number)\n\t}\n}\n\nfunc main() {\n\tvar N int\n\tfmt.Scan(&N)\n\tfor i := 0; i < N; i++ {\n\t\tvar temp int\n\t\tfmt.Scan(&temp)\n\t\tfizzBuzz(temp)\n\t}}",
		"problem": {
			"limit": 2000,
			"accuracy": -3,
			"tests": [
				{
					"name": "A",
					"in": "3 \n1 2 3\n",
					"out": "1\n2\nFizz\n"
				},
				{
					"name": "B",
					"in": "3\n2 5 1\n",
					"out": "2\nBuzz\n1\n"
				}
			]
		}
	}

	`))

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
