package tasks

import (
	gocontext "context"
	"time"

	"github.com/containerd/console"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// HandleConsoleResize resizes the console
func HandleConsoleResize(ctx gocontext.Context, task resizer, con console.Console) error {
	// do an initial resize of the console
	size, err := con.Size()
	if err != nil {
		return err
	}
	go func() {
		prevSize := size
		for {
			time.Sleep(time.Millisecond * 250)

			size, err := con.Size()
			if err != nil {
				log.G(ctx).WithError(err).Error("get pty size")
				continue
			}

			if size.Width != prevSize.Width || size.Height != prevSize.Height {
				if err := task.Resize(ctx, uint32(size.Width), uint32(size.Height)); err != nil {
					log.G(ctx).WithError(err).Error("resize pty")
				}
				prevSize = size
			}
		}
	}()
	return nil
}

// NewTask creates a new task
func NewTask(ctx gocontext.Context, client *containerd.Client, container containerd.Container, _ string, tty, nullIO bool, _fifoDir string, opts ...containerd.NewTaskOpts) (containerd.Task, error) {
	ioCreator := cio.NewCreator(cio.WithStdio)
	if tty {
		ioCreator = cio.NewCreator(cio.WithStdio, cio.WithTerminal)
	}
	if nullIO {
		if tty {
			return nil, errors.New("tty and null-io cannot be used together")
		}
		ioCreator = cio.NullIO
	}
	return container.NewTask(ctx, ioCreator)
}

func getNewTaskOpts(_ *cli.Context) []containerd.NewTaskOpts {
	return nil
}
