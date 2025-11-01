package service

import (
	"context"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
)

type Uploader interface {
	Upload(ctx context.Context, file io.Reader, folder string, publicID string) (string, error)
	Delete(ctx context.Context, publicID string) error
	GetClient() *cloudinary.Cloudinary
}