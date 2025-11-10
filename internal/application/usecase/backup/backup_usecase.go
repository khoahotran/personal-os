package backup

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type BackupUseCase struct {
	cfg      config.Config
	uploader service.Uploader
	logger   logger.Logger
}

func NewBackupUseCase(cfg config.Config, uploader service.Uploader, log logger.Logger) *BackupUseCase {
	return &BackupUseCase{
		cfg:      cfg,
		uploader: uploader,
		logger:   log,
	}
}

func (uc *BackupUseCase) Execute(ctx context.Context) {
	uc.logger.Info("Starting database backup...")

	dsn := uc.cfg.DB.DSN

	cmd := exec.Command("pg_dump", "--dbname="+dsn, "--format=c")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		uc.logger.Error("pg_dump failed", err, zap.String("stderr", stderr.String()))
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup-%s.dump", timestamp)
	folder := "backups/database"
	publicID := fmt.Sprintf("%s/%s", folder, filename)

	fileReader := bytes.NewReader(out.Bytes())

	uploadURL, err := uc.uploader.Upload(ctx, fileReader, folder, publicID)
	if err != nil {
		uc.logger.Error("Failed to upload backup to Cloudinary", err)
		return
	}

	uc.logger.Info("Database backup completed and uploaded successfully",
		zap.String("url", uploadURL),
		zap.String("public_id", publicID),
	)
}
