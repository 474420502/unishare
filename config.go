package unishare

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigServer struct {
	ServerID uint64   `yaml:"serverid"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Cluster  []string `yaml:"cluster"`
}

func (cfg *ConfigServer) Address() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func (cfg *ConfigServer) AddressWith(p int) string {
	return fmt.Sprintf("%s:%d", cfg.Host, p)
}

func LoadConfig(etcpath string) (*ConfigServer, error) {
	f, err := os.Open(etcpath)
	if err != nil {
		return nil, err
	}
	cfg := ConfigServer{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	return &cfg, err
}

func LoadAllConfig(etcpath string) (result []*ConfigServer) {

	filepath.WalkDir(etcpath, func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".yaml") {
			cfg, err := LoadConfig(path)
			if err != nil {
				panic(err)
			}
			result = append(result, cfg)
		}

		return nil
	})

	return result
}
