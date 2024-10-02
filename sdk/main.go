package main

import (
	"context"
	"fmt"
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	"time"
)

func main() {
	client, err := containerd.New("/var/run/containerd/containerd.sock")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "example")
	image, err := client.Pull(ctx, "docker.io/library/redis:alpine", containerd.WithPullUnpack)
	if err != nil {
		panic(err)
	}

	container, err := client.NewContainer(ctx, "redis-server", containerd.WithNewSnapshot("redis-server-snapshot", image), containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		panic(err)
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		panic(err)
	}
	defer task.Delete(ctx)

	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		panic(err)
	}

	if err := task.Start(ctx); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)
	if err := task.Kill(ctx, 9); err != nil {
		panic(err)
	}

	status := <-exitStatusC
	code, exitedAt, err := status.Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("redis-server exited with status", code, "at", exitedAt)
}
