package httpx_test

import (
	"context"
	"fmt"
	"log"

	"git.sr.ht/~jamesponddotco/httpx-go"
)

func ExampleClient_Get() {
	client := httpx.NewClientWithCache(nil)

	resp, err := client.Get(context.Background(), "https://example.com/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output: 200
}
