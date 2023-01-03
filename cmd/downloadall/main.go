package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/dillonstreator/dalle"
	"golang.org/x/sync/errgroup"
)

func main() {
	apiKey := os.Getenv("DALLE_API_KEY")
	if len(apiKey) == 0 {
		log.Fatal("DALLE_API_KEY required")
	}

	concurrency := 5

	concurrencyStr := os.Getenv("CONCURRENCY")
	if len(concurrencyStr) != 0 {
		var err error

		concurrency, err = strconv.Atoi(concurrencyStr)
		if err != nil {
			log.Fatalf("invalid `CONCURRENCY` value: %s", err.Error())
		}
	}

	imagesPath := os.Getenv("IMAGES_PATH")
	if len(imagesPath) == 0 {
		imagesPath = "images"
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

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(concurrency)

	for _, t := range res.Data {
		if egCtx.Err() != nil {
			break
		}

		if t.Status != dalle.StatusSucceeded {
			fmt.Printf("task id %s not completed yet.. skipping\n", t.ID)
			continue
		}

		for _, generation := range t.Generations.Data {

			g := generation
			eg.Go(func() error {

				reader, err := dalleClient.Download(egCtx, g.ID)
				if err != nil {
					return err
				}
				defer reader.Close()

				file, err := os.Create(path.Join(imagesPath, g.ID+".png"))
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.CopyBuffer(file, reader, make([]byte, 2048))
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
