package media_storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/config"
)

type cloudinaryAdapter struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryAdapter(cfg config.Config) (service.Uploader, error) {

	if cfg.Cloudinary.CloudName == "" {
		return nil, fmt.Errorf("cloudinary cloud_name has not config")
	}

	cld, err := cloudinary.NewFromParams(
		cfg.Cloudinary.CloudName,
		cfg.Cloudinary.ApiKey,
		cfg.Cloudinary.ApiSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot init cloudinary: %w", err)
	}

	log.Println("connect Cloudinary successfully.")
	return &cloudinaryAdapter{cld: cld}, nil
}

func (a *cloudinaryAdapter) Upload(ctx context.Context, file io.Reader, folder string, publicID string) (string, error) {
	uploadParams := uploader.UploadParams{
		PublicID: publicID,
		Folder:   folder,
	}
	result, err := a.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload cloudinary: %w", err)
	}
	return result.SecureURL, nil
}

func (a *cloudinaryAdapter) Delete(ctx context.Context, publicID string) error {
	_, err := a.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete cloudinary: %w", err)
	}
	return nil
}

func (a *cloudinaryAdapter) GetClient() *cloudinary.Cloudinary {
	return a.cld
}
