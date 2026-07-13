package services

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

type UploadedImage struct {
	URL      string
	PublicID string
}

func UploadAvatar(cfg CloudinaryConfig, userID string, dataURI string) (*UploadedImage, error) {
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}

	ctx := context.Background()
	publicID := fmt.Sprintf("seismic/avatars/%s", userID)
	overwrite := true

	result, err := cld.Upload.Upload(ctx, dataURI, uploader.UploadParams{
		PublicID:  publicID,
		Overwrite: &overwrite,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}
	if result.Error.Message != "" {
		return nil, fmt.Errorf("cloudinary upload failed: %s", result.Error.Message)
	}

	return &UploadedImage{URL: result.SecureURL, PublicID: result.PublicID}, nil
}

// DeleteAvatar removes an image from Cloudinary by its public ID.
// Used only when a user explicitly removes their avatar without
// choosing a replacement.
func DeleteAvatar(cfg CloudinaryConfig, publicID string) error {
	if publicID == "" {
		return nil
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}
