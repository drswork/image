// Transform shows how to load in an image and apply the transforms
// indicated in its metadata.
package main

import (
	"context"
	"github.com/drswork/image"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// Read in the image. Apply any rotation, gamma, and color
	// correction transformations that may be noted in the metadata.
	i, _, _, t, err := image.DecodeWithOptions(ctx, os.Stdin, image.ImageTransformOptions{
		RotationTransform: image.ForwardImageTransform,
		ColorTransform:    image.ForwardImageTransform,
		GammaTransform:    image.ForwardImageTransform,
	})

	if err != nil {
		log.Fatalf("Image read failed: %v", err)
	}
}
