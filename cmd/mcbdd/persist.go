package main

import (
	"encoding/json"
	"errors"
	"os"
)

func (d *Daemon) LoadFromDisk() error {
	f, err := os.OpenFile(d.statefile, os.O_RDONLY, 0o660)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&d.userTokens)
}

func (d *Daemon) SaveToDisk() error {
	f, err := os.OpenFile(d.statefile, os.O_CREATE|os.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(d.userTokens)
}
