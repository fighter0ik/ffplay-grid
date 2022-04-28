package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kbinani/screenshot"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	gw := flag.Int("w", 4, "grid gw in cells")
	gh := flag.Int("h", 4, "grid gh in cells")
	input := flag.String("i", "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mp4", "input")
	flag.Parse()

	dw, dh := displaySize()
	if dw < 1 || dh < 1 {
		panic("failed to get display size")
	}
	log.Printf("dw: %v. dh: %v", dw, dh)

	cw := dw / *gw
	ch := dh / *gh
	log.Printf("cw: %v. ch: %v", cw, ch)

	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	commands := make([]*exec.Cmd, 0, *gw**gh)
	for i := 0; i < *gw; i++ {
		for j := 0; j < *gh; j++ {
			cmd := exec.CommandContext(ctx, "ffplay",
				"-left", fmt.Sprintf("%d", i*cw),
				"-top", fmt.Sprintf("%d", j*ch),
				"-x", fmt.Sprintf("%d", cw),
				"-y", fmt.Sprintf("%d", ch),
				"-an",
				"-noborder",
				*input)
			log.Printf("i: %v. j: %v. cmd: %v", i, j, cmd.String())

			c := cellConsole{
				i: i,
				j: j,
			}

			cmd.Stdout = c
			cmd.Stderr = c

			commands = append(commands, cmd)
		}
	}
	for _, cmd := range commands {
		if err := cmd.Start(); err != nil {
			panic(err)
		}
	}

	sc := make(chan os.Signal)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	<-sc
}

func displaySize() (int, int) {
	bounds := screenshot.GetDisplayBounds(0)
	return bounds.Dx(), bounds.Dy()
}

type cellConsole struct {
	i int
	j int
}

func (c cellConsole) Write(p []byte) (n int, err error) {
	fmt.Printf("i: %v. j: %v. %s\n", c.i, c.j, p)
	return len(p), nil
}
