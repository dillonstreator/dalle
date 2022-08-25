# Dalle

go client for [DALL*E](https://openai.com/dall-e-2/)

## Usage

This tool requires access to DALL*E. If you don't currently have access, sign up for the [DALL*E wait list](https://labs.openai.com/waitlist).

Find the Bearer Token

- Go to https://labs.openai.com/
- Open Network Tab in Developer Tools
- Type a prompt and press "Generate"
- Look for the xhr request to https://labs.openai.com/api/labs/tasks
- In the request header, look for `Authorization` then get the Bearer Token which will act as the api key

> Note: Do not include "Bearer " in the api key

### Using the api key
```go
// create the client with the bearer token api key
dalleClient, err := dalle.NewHTTPClient("your-api-key-here")
// handle err

// generate a task to create an image with a prompt
task, err := dalleClient.Generate(ctx, "neon sports car driving into sunset, synthwave, cyberpunk")
// handle err

// poll the task.ID until status is succeeded
var t *dalle.Task
for {
    time.Sleep(time.Second * 3)

    t, err = dalleClient.GetTask(ctx, task.ID)
    // handle err

    if t.Status == dalle.StatusSucceeded {
        fmt.Println("task succeeded")
        break
    } else if t.Status == dalle.StatusRejected {
        log.Fatal("rejected: ", t.ID)
    }

    fmt.Println("task still pending")
}

// download the first generated image
reader, err := dalleClient.Download(ctx, t.Generations.Data[0].ID)
// handle err and close readCloser
```

## Examples

- [Generating, polling, and downloading a single image](./cmd/example/main.go)
- [Downloading all generated images from first 50 tasks](./cmd/downloadall/main.go)
