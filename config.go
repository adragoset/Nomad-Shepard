package shepard

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

//LoadShepardConfiguration loads configuration for the allocation watcher
func LoadShepardConfiguration() *Config {
	var config Config
	configFile, err := os.Open(path.Join(os.Getenv("SHEPARD_CONFIG_PATH"), "shepard.json"))
	defer configFile.Close()
	if err != nil {
		log.Fatalf("Could not find allocation_watcher.json configuration:%s", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	config.AllocWatcher.NomadAllocDir = path.Join(os.Getenv("NOMAD_WORKSPACE"), config.AllocWatcher.NomadAllocDir)
	return &config
}
