package pkgconfig

import (
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func IsValidMode(mode string) bool {
	switch mode {
	case
		ModeText,
		ModeJSON,
		ModeFlatJSON:
		return true
	}
	return false
}

type Config struct {
	Global               ConfigGlobal       `yaml:"global"`
	Collectors           ConfigCollectors   `yaml:"collectors"`
	IngoingTransformers  ConfigTransformers `yaml:"collectors-transformers"`
	Loggers              ConfigLoggers      `yaml:"loggers"`
	OutgoingTransformers ConfigTransformers `yaml:"loggers-transformers"`
	Multiplexer          ConfigMultiplexer  `yaml:"multiplexer"`
	Pipelines            []ConfigPipelines  `yaml:"pipelines"`
}

func (c *Config) SetDefault() {
	// Set default config for global part
	c.Global.SetDefault()

	// Set default config for multiplexer
	c.Multiplexer.SetDefault()

	// sSet default config for collectors
	c.Collectors.SetDefault()

	// transformers for collectors
	c.IngoingTransformers.SetDefault()

	// Set default config for loggers
	c.Loggers.SetDefault()

	// Transformers for loggers
	c.OutgoingTransformers.SetDefault()
}

func (c *Config) IsValid(userCfg map[string]interface{}) error {
	for userKey, userValue := range userCfg {
		switch userKey {
		case "global":
			if kvMap, ok := userValue.(map[string]interface{}); ok {
				if err := c.Global.Check(kvMap); err != nil {
					return errors.Errorf("global section - %s", err)
				}
			} else {
				return errors.Errorf("unexpected type for global value, got %T", kvMap)
			}

		case "multiplexer":
			if kvMap, ok := userValue.(map[string]interface{}); ok {
				if err := c.Multiplexer.IsValid(kvMap); err != nil {
					return errors.Errorf("multiplexer section - %s", err)
				}
			} else {
				return errors.Errorf("unexpected type for multiplexer value, got %T", kvMap)
			}

		case "pipelines":
			for i, cv := range userValue.([]interface{}) {
				cfg := ConfigPipelines{}
				if err := cfg.IsValid(cv.(map[string]interface{})); err != nil {
					return errors.Errorf("stanza(index=%d) - %s", i, err)
				}
			}

		default:
			return errors.Errorf("unknown key=%s\n", userKey)
		}
	}
	return nil
}

func (c *Config) GetServerIdentity() string {
	if len(c.Global.ServerIdentity) > 0 {
		return c.Global.ServerIdentity
	} else {
		hostname, err := os.Hostname()
		if err == nil {
			return hostname
		} else {
			return "undefined"
		}
	}
}

func GetDefaultConfig() *Config {
	config := &Config{}
	config.SetDefault()
	return config
}

func CheckConfigWithTags(v reflect.Value, userCfg map[string]interface{}) error {
	t := v.Type()
	for k, kv := range userCfg {
		keyExist := false
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := t.Field(i)

			// get name from yaml tag
			fieldTag := fieldType.Tag.Get("yaml")
			tagClean := strings.TrimSuffix(fieldTag, ",flow")

			// compare
			if tagClean == k {
				keyExist = true
			}
			if fieldValue.Kind() == reflect.Struct && tagClean == k {
				if kvMap, ok := kv.(map[string]interface{}); ok {
					err := CheckConfigWithTags(fieldValue, kvMap)
					if err != nil {
						return errors.Errorf("%s in subkey=`%s`", err, k)
					}
				} else {
					return errors.Errorf("unexpected type for key `%s`, got %T", k, kv)
				}
			}
		}

		if !keyExist {
			return errors.Errorf("unknown key=`%s`", k)
		}
	}
	return nil
}

func ReloadConfig(configPath string, config *Config) error {
	// Open config file
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil
	}
	defer configFile.Close()

	// Check config to detect unknown keywords
	if err := CheckConfig(configFile); err != nil {
		return err
	}

	// Init new YAML decode
	configFile.Seek(0, 0)
	d := yaml.NewDecoder(configFile)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return err
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	// Open config file
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	// Check config to detect unknown keywords
	if err := CheckConfig(configFile); err != nil {
		return nil, err
	}

	// Init new YAML decode
	configFile.Seek(0, 0)
	d := yaml.NewDecoder(configFile)

	// Start YAML decoding to go
	config := &Config{}
	config.SetDefault()

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func CheckConfig(configFile *os.File) error {
	// Read config file bytes
	configBytes, err := io.ReadAll(configFile)
	if err != nil {
		return errors.Wrap(err, "Error reading configuration file")
	}

	// Unmarshal YAML to map
	userCfg := make(map[string]interface{})
	err = yaml.Unmarshal(configBytes, &userCfg)
	if err != nil {
		return errors.Wrap(err, "error parsing YAML file")
	}

	// check the user config with the default one
	config := &Config{}
	config.SetDefault()

	// check if the provided config is valid
	return config.IsValid(userCfg)
}
