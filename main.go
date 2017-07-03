package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"

	"os"

	"time"

	"github.com/fsnotify/fsnotify"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func eventWarrentsReload(event fsnotify.Event) bool {
	return event.Op == fsnotify.Write || event.Op == fsnotify.Rename || event.Op == fsnotify.Remove
}

func runLove(projectDir string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "love", projectDir)
	cmd.Stdout = os.Stdout

	must(cmd.Start())
	return cancel
}

func main() {
	logger := log.New(os.Stdout, "SERVER::", 0)

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		// TODO print usage
		os.Exit(1)
	}

	dir := args[0]
	cancel := runLove(dir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if eventWarrentsReload(event) {
					cancel()
					fmt.Println("=====================")
					cancel = runLove(dir)
				}
			case err := <-watcher.Errors:
				logger.Println("error:", err)
			}
		}
	}()

	s := NewServer(logger)
	err = s.Listen(context.Background(), "localhost:8080")
	if err != nil {
		logger.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	go func() {
		for {
			time.Sleep(time.Second * 3)
			s.Command(CmdRestart)
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		logger.Fatal(err)
	}
	<-done
}
