package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/urfave/cli"
	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

var containerCommand = cli.Command{
	Name:    "container",
	Aliases: []string{"ctr"},
	Subcommands: []cli.Command{
		createContainerCommand,
		startContainerCommand,
		stopContainerCommand,
		removeContainerCommand,
		containerStatusCommand,
		listContainersCommand,
		execSyncCommand,
	},
}

type createOptions struct {
	// configPath is path to the config for container
	configPath string
	// name sets the container name
	name string
	// podID of the container
	podID string
	// labels for the container
	labels map[string]string
}

var createContainerCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "pod",
			Usage: "the id of the pod sandbox to which the container belongs",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "config.json",
			Usage: "the path of a container config file",
		},
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "the name of the container",
		},
		cli.StringSliceFlag{
			Name:  "label",
			Usage: "add key=value labels to the container",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		if !context.IsSet("pod") {
			return fmt.Errorf("Please specify the id of the pod sandbox to which the container belongs via the --pod option")
		}

		opts := createOptions{
			configPath: context.String("config"),
			name:       context.String("name"),
			podID:      context.String("pod"),
			labels:     make(map[string]string),
		}

		for _, l := range context.StringSlice("label") {
			pair := strings.Split(l, "=")
			if len(pair) != 2 {
				return fmt.Errorf("incorrectly specified label: %v", l)
			}
			opts.labels[pair[0]] = pair[1]
		}

		// Test RuntimeServiceClient.CreateContainer
		err = CreateContainer(client, opts)
		if err != nil {
			return fmt.Errorf("Creating container failed: %v", err)
		}
		return nil
	},
}

var startContainerCommand = cli.Command{
	Name:  "start",
	Usage: "start a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the container",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = StartContainer(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("Starting the container failed: %v", err)
		}
		return nil
	},
}

var stopContainerCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the container",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = StopContainer(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("Stopping the container failed: %v", err)
		}
		return nil
	},
}

var removeContainerCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the container",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = RemoveContainer(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("Removing the container failed: %v", err)
		}
		return nil
	},
}

var containerStatusCommand = cli.Command{
	Name:  "status",
	Usage: "get the status of a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the container",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = ContainerStatus(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("Getting the status of the container failed: %v", err)
		}
		return nil
	},
}

var execSyncCommand = cli.Command{
	Name:  "execsync",
	Usage: "exec a command synchronously in a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the container",
		},
		cli.Int64Flag{
			Name:  "timeout",
			Value: 0,
			Usage: "timeout for the command",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = ExecSync(client, context.String("id"), context.Args(), context.Int64("timeout"))
		if err != nil {
			return fmt.Errorf("execing command in container failed: %v", err)
		}
		return nil
	},
}

type listOptions struct {
	// id of the container
	id string
	// podID of the container
	podID string
	// state of the container
	state string
	// quiet is for listing just container IDs
	quiet bool
	// labels are selectors for the container
	labels map[string]string
}

var listContainersCommand = cli.Command{
	Name:  "list",
	Usage: "list containers",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "list only container IDs",
		},
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "filter by container id",
		},
		cli.StringFlag{
			Name:  "pod",
			Value: "",
			Usage: "filter by container pod id",
		},
		cli.StringFlag{
			Name:  "state",
			Value: "",
			Usage: "filter by container state",
		},
		cli.StringSliceFlag{
			Name:  "label",
			Usage: "filter by key=value label",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)
		opts := listOptions{
			id:     context.String("id"),
			podID:  context.String("pod"),
			state:  context.String("state"),
			quiet:  context.Bool("quiet"),
			labels: make(map[string]string),
		}

		for _, l := range context.StringSlice("label") {
			pair := strings.Split(l, "=")
			if len(pair) != 2 {
				return fmt.Errorf("incorrectly specified label: %v", l)
			}
			opts.labels[pair[0]] = pair[1]
		}

		err = ListContainers(client, opts)
		if err != nil {
			return fmt.Errorf("listing containers failed: %v", err)
		}
		return nil
	},
}

// CreateContainer sends a CreateContainerRequest to the server, and parses
// the returned CreateContainerResponse.
func CreateContainer(client pb.RuntimeServiceClient, opts createOptions) error {
	config, err := loadContainerConfig(opts.configPath)
	if err != nil {
		return err
	}

	// Override the name by the one specified through CLI
	if opts.name != "" {
		config.Metadata.Name = &opts.name
	}

	for k, v := range opts.labels {
		config.Labels[k] = v
	}

	r, err := client.CreateContainer(context.Background(), &pb.CreateContainerRequest{
		PodSandboxId: &opts.podID,
		Config:       config,
	})
	if err != nil {
		return err
	}
	fmt.Println(*r.ContainerId)
	return nil
}

// StartContainer sends a StartContainerRequest to the server, and parses
// the returned StartContainerResponse.
func StartContainer(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := client.StartContainer(context.Background(), &pb.StartContainerRequest{
		ContainerId: &ID,
	})
	if err != nil {
		return err
	}
	fmt.Println(ID)
	return nil
}

// StopContainer sends a StopContainerRequest to the server, and parses
// the returned StopContainerResponse.
func StopContainer(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := client.StopContainer(context.Background(), &pb.StopContainerRequest{
		ContainerId: &ID,
	})
	if err != nil {
		return err
	}
	fmt.Println(ID)
	return nil
}

// RemoveContainer sends a RemoveContainerRequest to the server, and parses
// the returned RemoveContainerResponse.
func RemoveContainer(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := client.RemoveContainer(context.Background(), &pb.RemoveContainerRequest{
		ContainerId: &ID,
	})
	if err != nil {
		return err
	}
	fmt.Println(ID)
	return nil
}

// ContainerStatus sends a ContainerStatusRequest to the server, and parses
// the returned ContainerStatusResponse.
func ContainerStatus(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	r, err := client.ContainerStatus(context.Background(), &pb.ContainerStatusRequest{
		ContainerId: &ID})
	if err != nil {
		return err
	}
	fmt.Printf("ID: %s\n", *r.Status.Id)
	if r.Status.Metadata != nil {
		if r.Status.Metadata.Name != nil {
			fmt.Printf("Name: %s\n", *r.Status.Metadata.Name)
		}
		if r.Status.Metadata.Attempt != nil {
			fmt.Printf("Attempt: %v\n", *r.Status.Metadata.Attempt)
		}
	}
	if r.Status.State != nil {
		fmt.Printf("Status: %s\n", r.Status.State)
	}
	if r.Status.CreatedAt != nil {
		ctm := time.Unix(0, *r.Status.CreatedAt)
		fmt.Printf("Created: %v\n", ctm)
	}
	if r.Status.StartedAt != nil {
		stm := time.Unix(0, *r.Status.StartedAt)
		fmt.Printf("Started: %v\n", stm)
	}
	if r.Status.FinishedAt != nil {
		ftm := time.Unix(0, *r.Status.FinishedAt)
		fmt.Printf("Finished: %v\n", ftm)
	}
	if r.Status.ExitCode != nil {
		fmt.Printf("Exit Code: %v\n", *r.Status.ExitCode)
	}

	return nil
}

// ExecSync sends an ExecSyncRequest to the server, and parses
// the returned ExecSyncResponse.
func ExecSync(client pb.RuntimeServiceClient, ID string, cmd []string, timeout int64) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	r, err := client.ExecSync(context.Background(), &pb.ExecSyncRequest{
		ContainerId: &ID,
		Cmd:         cmd,
		Timeout:     &timeout,
	})
	if err != nil {
		return err
	}
	fmt.Println("Stdout:")
	fmt.Println(string(r.Stdout))
	fmt.Println("Stderr:")
	fmt.Println(string(r.Stderr))
	fmt.Printf("Exit code: %v\n", *r.ExitCode)

	return nil
}

// ListContainers sends a ListContainerRequest to the server, and parses
// the returned ListContainerResponse.
func ListContainers(client pb.RuntimeServiceClient, opts listOptions) error {
	filter := &pb.ContainerFilter{}
	if opts.id != "" {
		filter.Id = &opts.id
	}
	if opts.podID != "" {
		filter.PodSandboxId = &opts.podID
	}
	if opts.state != "" {
		st := pb.ContainerState_CONTAINER_UNKNOWN
		switch opts.state {
		case "created":
			st = pb.ContainerState_CONTAINER_CREATED
			filter.State = &st
		case "running":
			st = pb.ContainerState_CONTAINER_RUNNING
			filter.State = &st
		case "stopped":
			st = pb.ContainerState_CONTAINER_EXITED
			filter.State = &st
		default:
			log.Fatalf("--state should be one of created, running or stopped")
		}
	}
	if opts.labels != nil {
		filter.LabelSelector = opts.labels
	}
	r, err := client.ListContainers(context.Background(), &pb.ListContainersRequest{
		Filter: filter,
	})
	if err != nil {
		return err
	}
	for _, c := range r.GetContainers() {
		if opts.quiet {
			fmt.Println(*c.Id)
			continue
		}
		fmt.Printf("ID: %s\n", *c.Id)
		fmt.Printf("Pod: %s\n", *c.PodSandboxId)
		if c.Metadata != nil {
			if c.Metadata.Name != nil {
				fmt.Printf("Name: %s\n", *c.Metadata.Name)
			}
			if c.Metadata.Attempt != nil {
				fmt.Printf("Attempt: %v\n", *c.Metadata.Attempt)
			}
		}
		if c.State != nil {
			fmt.Printf("Status: %s\n", *c.State)
		}
		if c.CreatedAt != nil {
			ctm := time.Unix(0, *c.CreatedAt)
			fmt.Printf("Created: %v\n", ctm)
		}
		if c.Labels != nil {
			fmt.Println("Labels:")
			for _, k := range getSortedKeys(c.Labels) {
				fmt.Printf("\t%s -> %s\n", k, c.Labels[k])
			}
		}
		if c.Annotations != nil {
			fmt.Println("Annotations:")
			for _, k := range getSortedKeys(c.Annotations) {
				fmt.Printf("\t%s -> %s\n", k, c.Annotations[k])
			}
		}
		fmt.Println()
	}
	return nil
}
