package utils

import (
	"context"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"

	"log"
)

var BlobBucket *blob.Bucket
var Ctx context.Context

func InitBlobBucket() {
	Ctx = context.Background()
	BlobBucket, err := blob.OpenBucket(Ctx, "mem://")
	if err != nil {
		log.Fatal(err)
	}
	defer BlobBucket.Close()
}
