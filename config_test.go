package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
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

//nolint:gocognit
func TestParseConfig(t *testing.T) {
	t.Parallel()

	//nolint:paralleltest
	t.Run("default values", func(t *testing.T) {
		var actualEnvs = map[string]string{
			EnvPrefix + "MONGO_URI":      "",
			EnvPrefix + "MONGO_DATABASE": "",
			EnvPrefix + "CHATGPT_TOKEN":  "",
		}

		setEnv(actualEnvs)

		cfg, err := ParseConfig([]string{"foo", string(CreateUser)})
		if err != nil {
			t.Errorf("unexpected error %s", err)
		}

		expectedCfg := configWithDefaults(CreateUser)
		if diff := cmp.Diff(expectedCfg, cfg); diff != "" {
			t.Errorf("unexpected config (-want +got):\n%s", diff)
		}
	})

	//nolint:paralleltest
	t.Run("env values", func(t *testing.T) {
		var actualEnvs = map[string]string{
			EnvPrefix + "MONGO_URI":      "foobar",
			EnvPrefix + "MONGO_DATABASE": "bazbaz",
			EnvPrefix + "CHATGPT_TOKEN":  "atata",
		}

		setEnv(actualEnvs)

		cfg, err := ParseConfig([]string{"foo", string(CreateUser)})
		if err != nil {
			t.Errorf("unexpected error %s", err)
		}

		expectedCfg := configWithDefaults(CreateUser)
		expectedCfg.Mongo.URI = "foobar"
		expectedCfg.Mongo.DatabaseName = "bazbaz"
		expectedCfg.ChatGPT.APIToken = "atata"
		if diff := cmp.Diff(expectedCfg, cfg); diff != "" {
			t.Errorf("unexpected config (-want +got):\n%s", diff)
		}
	})

	//nolint:paralleltest
	t.Run("cli add-word values", func(t *testing.T) {
		var actualEnvs = map[string]string{
			EnvPrefix + "MONGO_URI":      "",
			EnvPrefix + "MONGO_DATABASE": "",
			EnvPrefix + "CHATGPT_TOKEN":  "",
		}

		setEnv(actualEnvs)

		cfg, err := ParseConfig([]string{"foo", string(AddWord), "-user-id=abc", "-spelling=sss", "-definition=ddd", "-language=en_GB", "-lexical-category=adverb"})
		if err != nil {
			t.Errorf("unexpected error %s", err)
		}

		expectedCfg := configWithDefaults(AddWord)
		expectedCfg.UserID = "abc"
		expectedCfg.Spelling = "sss"
		expectedCfg.Definition = "ddd"
		expectedCfg.Language = "en_GB"
		expectedCfg.LexicalCategory = "adverb"

		if diff := cmp.Diff(expectedCfg, cfg); diff != "" {
			t.Errorf("unexpected config (-want +got):\n%s", diff)
		}
	})
}
