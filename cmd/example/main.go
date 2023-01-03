package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dillonstreator/dalle"
	"golang.org/x/sync/errgroup"
)

func main() {

	apiKey := os.Getenv("DALLE_API_KEY")
	if len(apiKey) == 0 {
		log.Fatal("DALLE_API_KEY required")
	}

	dalleClient, err := dalle.NewHTTPClient(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	task, err := dalleClient.Generate(ctx, "neon sports car driving into sunset, synthwave, cyberpunk")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("task id %s\n", task.ID)

	var t *dalle.Task
	for {
		time.Sleep(time.Second * 3)

		t, err = dalleClient.GetTask(ctx, task.ID)
		if err != nil {
			log.Fatal(err)
		}

		if t.Status == dalle.StatusSucceeded {
			fmt.Println("task succeeded")
			break
		} else if t.Status == dalle.StatusRejected {
			log.Fatal("rejected: ", t.ID)
		}

		fmt.Println("task still pending")
	}

	fmt.Printf("%d images generated", len(t.Generations.Data))

	eg := errgroup.Group{}

	for _, data := range t.Generations.Data {

		d := data

		eg.Go(func() error {

			reader, err := dalleClient.Download(ctx, d.ID)
			if err != nil {
				return err
			}
			defer reader.Close()

			file, err := os.Create(fmt.Sprintf("%s.png", d.ID))
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.CopyBuffer(file, reader, make([]byte, 512))
			if err != nil {
				return err
			}

			return nil
		})

	}

	err = eg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
