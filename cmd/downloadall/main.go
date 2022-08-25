package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

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

	res, err := dalleClient.ListTasks(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	eg := &errgroup.Group{}

	for _, t := range res.Data {

		if t.Status != dalle.StatusSucceeded {
			fmt.Printf("task id %s not completed yet.. skipping\n", t.ID)
			continue
		}

		for _, generation := range t.Generations.Data {

			g := generation
			eg.Go(func() error {

				reader, err := dalleClient.Download(ctx, g.ID)
				if err != nil {
					return err
				}
				defer reader.Close()

				file, err := os.Create("images/" + g.ID + ".png")
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(file, reader)
				if err != nil {
					return err
				}

				return nil
			})

		}

	}

	err = eg.Wait()
	if err != nil {
		log.Fatal(err)
	}

}
