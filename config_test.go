package main

import (
	"os"
	"reflect"
	"testing"
)

func setEnv(e map[string]string) {
	for k, v := range e {
		if v == "" {
			os.Unsetenv(k)
			continue
		}
		os.Setenv(k, v)
	}
}

func TestParseConfig(t *testing.T) {
	t.Parallel()

	//nolint:paralleltest
	t.Run("default values", func(t *testing.T) {
		var actualEnvs = map[string]string{
			EnvPrefix + "MONGO_URI":      "",
			EnvPrefix + "MONGO_DATABASE": "",
		}

		setEnv(actualEnvs)

		cfg, err := ParseConfig([]string{"foo", string(CreateUser)})
		if err != nil {
			t.Errorf("unexpected error %s", err)
		}

		expectedCfg := configWithDefaults(CreateUser)
		if !reflect.DeepEqual(expectedCfg, cfg) {
			t.Errorf("unexpected config. \nExpected: %#v\nGot: %#v", expectedCfg, cfg)
		}
	})

	//nolint:paralleltest
	t.Run("env values", func(t *testing.T) {
		var actualEnvs = map[string]string{
			EnvPrefix + "MONGO_URI":      "foobar",
			EnvPrefix + "MONGO_DATABASE": "bazbaz",
		}

		setEnv(actualEnvs)

		cfg, err := ParseConfig([]string{"foo", string(CreateUser)})
		if err != nil {
			t.Errorf("unexpected error %s", err)
		}

		expectedCfg := configWithDefaults(CreateUser)
		expectedCfg.Mongo.URI = "foobar"
		expectedCfg.Mongo.DatabaseName = "bazbaz"
		if !reflect.DeepEqual(expectedCfg, cfg) {
			t.Errorf("unexpected config. \nExpected: %#v\nGot: %#v", expectedCfg, cfg)
		}
	})
}
