// +build !windows

package tasks

import (
	gocontext "context"
	"os"
	"os/signal"

	"github.com/containerd/console"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
)

func init() {
	startCommand.Flags = append(startCommand.Flags, cli.BoolFlag{
		Name:  "no-pivot",
		Usage: "disable use of pivot-root (linux only)",
	})
}

// HandleConsoleResize resizes the console
func HandleConsoleResize(ctx gocontext.Context, task resizer, con console.Console) error {
	// do an initial resize of the console
	size, err := con.Size()
	if err != nil {
		return err
	}
	if err := task.Resize(ctx, uint32(size.Width), uint32(size.Height)); err != nil {
		log.G(ctx).WithError(err).Error("resize pty")
	}
	s := make(chan os.Signal, 16)
	signal.Notify(s, unix.SIGWINCH)
	go func() {
		for range s {
			size, err := con.Size()
			if err != nil {
				log.G(ctx).WithError(err).Error("get pty size")
				continue
			}
			if err := task.Resize(ctx, uint32(size.Width), uint32(size.Height)); err != nil {
				log.G(ctx).WithError(err).Error("resize pty")
			}
		}
	}()
	return nil
}

// NewTask creates a new task
func NewTask(ctx gocontext.Context, client *containerd.Client, container containerd.Container, checkpoint string, tty, nullIO bool, fifoDir string, opts ...containerd.NewTaskOpts) (containerd.Task, error) {
	stdio := cio.NewCreator(cio.WithStdio, cio.WithFIFODir(fifoDir))
	if checkpoint == "" {
		ioCreator := stdio
		if tty {
			ioCreator = cio.NewCreator(cio.WithStdio, cio.WithTerminal, cio.WithFIFODir(fifoDir))
		}
		if nullIO {
			if tty {
				return nil, errors.New("tty and null-io cannot be used together")
			}
			ioCreator = cio.NullIO
		}
		return container.NewTask(ctx, ioCreator, opts...)
	}
	im, err := client.GetImage(ctx, checkpoint)
	if err != nil {
		return nil, err
	}
	opts = append(opts, containerd.WithTaskCheckpoint(im))
	return container.NewTask(ctx, stdio, opts...)
}

func getNewTaskOpts(context *cli.Context) []containerd.NewTaskOpts {
	if context.Bool("no-pivot") {
		return []containerd.NewTaskOpts{containerd.WithNoPivotRoot}
	}
	return nil
}
