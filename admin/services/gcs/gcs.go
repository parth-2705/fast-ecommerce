package GCS

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/tryamigo/themis"
)

var ProjectID string
var BucketName string

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var Uploader *ClientUploader

func Init() error {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "admin/creds.json") // FILL IN WITH YOUR FILE PATH

	BucketName, err := themis.GetSecret("GCS_BUCKET_NAME")
	if err != nil {
		return err
	}
	ProjectID, err = themis.GetSecret("GCS_PROJECT_ID")
	if err != nil {
		return err
	}

	uploadPath, err := themis.GetSecret("GCS_UPLOAD_FOLDER")
	if err != nil {
		return err
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	Uploader = &ClientUploader{
		cl:         client,
		bucketName: BucketName,
		projectID:  ProjectID,
		uploadPath: uploadPath + "/",
	}

	return nil
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	// Upload an object with storage.Writer.

	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
