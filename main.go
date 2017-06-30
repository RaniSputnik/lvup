package main

import (
	"context"
	"flag"
	"log"
	"os/exec"

	"os"

	"fmt"

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
	must(cmd.Start())
	return cancel
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		// TODO print usage
		os.Exit(1)
	}

	dir := args[0]
	fmt.Printf("%s", dir)
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
					cancel = runLove(dir)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
