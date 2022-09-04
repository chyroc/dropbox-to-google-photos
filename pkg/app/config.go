package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Account      string `json:"account"`
	GooglePhotos struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"google_photos"`
	Dropbox struct {
		Token   string `json:"token"`
		RootDir string `json:"root_dir"`
	} `json:"dropbox"`
	Worker int `json:"worker"`
}

func (r *App) loadConfig() error {
	bs, err := ioutil.ReadFile(r.configPath())
	if err != nil {
		return err
	}
	cfg := new(Config)
	err = json.Unmarshal(bs, cfg)
	if err != nil {
		return err
	}
	r.config = cfg

	if r.config.Worker == 0 {
		r.config.Worker = 1
	}
	return nil
}

func (r *App) InitConfig(force bool) error {
	if !force {
		bs, _ := ioutil.ReadFile(r.configPath())
		if len(bs) > 0 {
			return fmt.Errorf("config file already exists, use --force to overwrite")
		}
	}

	bs, err := json.MarshalIndent(&Config{
		Account: "",
		GooglePhotos: struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}{
			ClientID:     "google client id",
			ClientSecret: "google client secret",
		},
		Dropbox: struct {
			Token   string `json:"token"`
			RootDir string `json:"root_dir"`
		}{
			Token:   "dropbox token",
			RootDir: "dropbox sync root dir",
		},
		Worker: 5,
	}, "", "  ")
	if err != nil {
		return err
	}
	err = os.MkdirAll(r.workDir, 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(r.configPath(), bs, 0644)
}

func (r *App) configPath() string {
	return r.workDir + "/config.json"
}
