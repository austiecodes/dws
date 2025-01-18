package services

import (
	"fmt"
	"testing"

	"github.com/austiecodes/dws/start"
	"github.com/gin-gonic/gin"
)

func TestListImage(t *testing.T) {
	start.MustInit()

	ctx := &gin.Context{}
	imgs, err := ListImages(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(imgs)
}
