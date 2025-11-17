package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func (d *Daemon) loadState() error {
	stateVer := struct {
		Version int `json:"version"`
	}{}
	if err := d.loadFromDisk(&stateVer); err != nil {
		return fmt.Errorf("cant detect state version: %w", err)
	}
	switch stateVer.Version {
	case 0:
		slog.Warn("loading old state version", "stateVer", stateVer.Version)
		if err := d.loadFromDisk(&d.userTokens); err != nil {
			return fmt.Errorf("cant load state v%d: %w", stateVer.Version, err)
		}
		d.stateUnsaved = true
	case 1:
		state := struct {
			Version    int               `json:"version"`
			UserTokens map[string]string `json:"userTokens"`
		}{}
		if err := d.loadFromDisk(&state); err != nil {
			return fmt.Errorf("cant load state v%d: %w", stateVer.Version, err)
		}
		for k, v := range state.UserTokens {
			dec, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return fmt.Errorf("cant decode pass from %s: %w", k, err)
			}
			d.userTokens[k] = string(dec)
		}
	}
	return nil
}

func (d *Daemon) saveState() error {
	encTokens := make(map[string]string, len(d.userTokens))
	for k, v := range d.userTokens {
		encTokens[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	state := struct {
		Version    int               `json:"version"`
		UserTokens map[string]string `json:"userTokens"`
	}{
		Version:    1,
		UserTokens: encTokens,
	}
	return d.saveToDisk(state)
}

func (d *Daemon) loadFromDisk(state any) error {
	f, err := os.OpenFile(d.stateFilepath, os.O_RDONLY, 0o660)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(state)
}

func (d *Daemon) saveToDisk(state any) error {
	f, err := os.OpenFile(d.stateFilepath, os.O_CREATE|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(state)
}
