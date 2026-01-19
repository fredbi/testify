package config

import (
	"io"
	"os"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"go.yaml.in/yaml/v3"
)

func TestLoad(t *testing.T) {
}

func TestLoadDefault(t *testing.T) {
	cfg, err := loadDefaults()
	if err != nil {
		t.Fatalf("cannot load defaults: %v", err)
	}

	err = dumpConfig(os.Stdout, cfg)
	if err != nil {
		t.Fatalf("cannot marshal to yaml: %v", err)
	}
}

func dumpConfig(w io.Writer, cfg *Config) error {
	var raw any
	err := mapstructure.Decode(cfg, &raw)
	if err != nil {
		return err
	}

	enc := yaml.NewEncoder(w)

	return enc.Encode(raw)
}
