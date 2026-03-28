package cmd

import (
	"fmt"

	"github.com/Ramyprojs/goclip/internal/config"
	"github.com/Ramyprojs/goclip/internal/db"
)

var appConfig = config.DefaultConfig()

func loadAppConfig() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	appConfig = cfg
	return nil
}

func loadAppConfigWithFallback(allowFallback bool) error {
	if err := loadAppConfig(); err != nil {
		if !allowFallback {
			return err
		}

		appConfig = config.DefaultConfig()
		return nil
	}

	return nil
}

func openStore() (*db.Store, error) {
	return db.OpenDB(appConfig.DBPath)
}

func configuredPreviewLength() int {
	if appConfig.PreviewLength > 0 {
		return appConfig.PreviewLength
	}

	return config.DefaultConfig().PreviewLength
}

func enforceMaxHistory(store *db.Store) error {
	if appConfig.MaxHistory <= 0 {
		return nil
	}

	clips, err := store.GetAllClips()
	if err != nil {
		return fmt.Errorf("load clips for max history: %w", err)
	}

	for i := appConfig.MaxHistory; i < len(clips); i++ {
		if err := store.DeleteClip(clips[i].ID); err != nil {
			return fmt.Errorf("trim history: %w", err)
		}
	}

	return nil
}
