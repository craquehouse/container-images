package main

import (
	"context"
	"testing"

	"github.com/craquehouse/container-images/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()
	image := testhelpers.GetTestImage("ghcr.io/craquehouse/gluetun-qb-port-sync:rolling")

	t.Run("Check /app/script.sh exists", func(t *testing.T) {
		testhelpers.TestFileExists(t, ctx, image, "/app/script.sh", nil)
	})
}
