package media_storage

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type cloudinaryAdapter struct {
	cld    *cloudinary.Cloudinary
	logger logger.Logger
}

func NewCloudinaryAdapter(cfg config.Config, log logger.Logger) (service.Uploader, error) {

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

	log.Info("Connect Cloudinary successfully.")
	return &cloudinaryAdapter{cld: cld, logger: log}, nil
}

func (a *cloudinaryAdapter) Upload(ctx context.Context, file io.Reader, folder string, publicID string) (string, error) {
	uploadParams := uploader.UploadParams{
		PublicID: publicID,
		Folder:   folder,
	}
	result, err := a.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", apperror.NewInternal("failed to upload to cloudinary", err)
	}
	return result.SecureURL, nil
}

func (a *cloudinaryAdapter) Delete(ctx context.Context, publicID string) error {
	_, err := a.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		a.logger.Warn("Failed to delete asset from cloudinary", zap.String("public_id", publicID), zap.Error(err))
	}
	return nil
}

func (a *cloudinaryAdapter) GetClient() *cloudinary.Cloudinary {
	return a.cld
}
