package config

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

func initConfig(filename string) *config {
	// find config file path
	_, currentFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(currentFile), "..", "..", filename)
	// read configs
	var config config
	if err := cleanenv.ReadConfig(path, &config); err != nil {
		log.Panicf("config.initConfig: failed loading configs: %s", err)
	}
	return &config
}
