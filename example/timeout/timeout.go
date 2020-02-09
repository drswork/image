// This program demonstrates reading an image file with a context that
// has a timeout set.
package main

import (
	"context"
	"fmt"
	"github.com/drswork/image"
	"log"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	// Read the image with our 3ms timeout. If it takes longer than that
	// to read then the decode will fail with an error.
	_, _, _, t, err := image.DecodeWithOptions(ctx, os.Stdin)
	if err != nil {
		log.Fatalf("Image read failed: %v", err)
	}
	fmt.Printf("Image is of type %v\n", t)
}
