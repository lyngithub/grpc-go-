package conf

import (
	"encoding/json"
	"log"
	"os"
)

var GlobalConfig *ProjectConfig

const configFile = "usergrowth.json"

type ProjectConfig struct {
	Db struct {
		Engine          string
		Username        string
		Password        string
		Host            string
		Port            int
		Database        string
		Charset         string
		ShowSql         bool
		MaxIdleConns    int
		MaxOpenConns    int
		ConnMaxLifetime int
	}
	Cache struct{}
}

func LoadConfigs() {
	LoadConfigFromFile(configFile)
}

func LoadConfigFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("conf.LoadConfigFromFile(%s) error=%s\n", filename, err.Error())
		return
	}
	defer file.Close()

	pc := &ProjectConfig{}
	if err := json.NewDecoder(file).Decode(pc); err != nil {
		log.Fatalf("conf.LoadConfigFromFile(%s) error=%s\n", filename, err.Error())
		return
	}

	if pc.Db.Username == "" {
		log.Fatalln("empty username in config ", filename)
		return
	}

	GlobalConfig = pc
}
