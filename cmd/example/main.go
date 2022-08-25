package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dillonstreator/dalle"
)

func main() {

	apiKey := os.Getenv("API_KEY")
	if len(apiKey) == 0 {
		log.Fatal("API_KEY required")
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

		if t.Status == "succeeded" {
			fmt.Println("task succeeded")
			break
		} else if t.Status == "rejected" {
			log.Fatal("rejected: ", t.ID)
		}

		fmt.Println("task still pending")
	}

	fmt.Printf("%d images generated", len(t.Generations.Data))

	reader, err := dalleClient.Download(ctx, t.Generations.Data[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	b, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile(fmt.Sprintf("%s.png", t.Generations.Data[0].ID), b, os.ModePerm)

}
