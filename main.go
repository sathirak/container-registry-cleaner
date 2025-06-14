package main

import (
	"fmt"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	val := githubactions.GetInput("val")
	if val == "" {
		githubactions.Fatalf("missing 'val'")
	}

	fmt.Println("Hello, world!")
}
