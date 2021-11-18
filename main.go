package main

import (
	"bytes"
	"flag"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

type Config struct {
	Glacier       Glacier
	PathsToBackup []string `yaml:"paths_to_backup"`
	Schedule      string
}

type Glacier struct {
	Id        string
	Secret    string
	Region    string
	VaultName string `yaml:"vault_name"`
}

func main() {
	var cfg Config
	var ConfigFile = flag.String("config_file", "config/dev.yaml", "Path to the YAML file containing the configuration.")
	flag.Parse()
	err := getConfigFromPath(*ConfigFile, &cfg)
	if err != nil {
		log.Fatalf("Failed to read the conf file %s: %v", *ConfigFile, err)
	}
	log.Println("start")

	db := NewFileDB()
	archiver := NewArchiver(cfg)
	uploader := NewUploader(cfg, db)
	scheduler := cron.New()

	_, err = scheduler.AddFunc(cfg.Schedule, func() {
		log.Println("run")
		var buf bytes.Buffer
		err := archiver.Create(&buf)
		if err != nil {
			log.Println("Unable to create archive", err)
		}
		err = uploader.Upload(bytes.NewReader(buf.Bytes()))
		if err != nil {
			log.Println("Unable to upload archive", err)
		}
		log.Println("done")
	})
	if err != nil {
		log.Fatalf("Unable to schedule upload: %v", err)
	}
	scheduler.Start()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}

func getConfigFromPath(configFilePath string, config interface{}) error {
	configRaw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	if err = yaml.UnmarshalStrict(configRaw, config); err != nil {
		return err
	}
	return nil
}
