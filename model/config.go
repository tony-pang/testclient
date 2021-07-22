package model

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

type Request struct {
	Body string `yaml:"body"`
}

type ResponseFromAlias struct {
	ID       string   `yaml:"id"`
	Messages []string `yaml:"messages"`
}

type User struct {
	ID      string              `yaml:"id"`
	Aliases []ResponseFromAlias `yaml:"aliases"`
}

type Test struct {
	Name     string    `yaml:"name"`
	Requests []Request `yaml:"requests"`
	Expected []User    `yaml:"expected"`
}

type Config struct {
	ProjectID       string        `yaml:"project_id"`
	ProjectIDHeader string        `yaml:"project_id_header"`
	TokenURL        string        `yaml:"token_url"`
	TestTimeout     time.Duration `yaml:"test_timeout"`
	DoormanURL      string        `yaml:"doorman_url"`
	TestServiceURL  string        `yaml:"test_service_url"`
	Tests           []Test        `yaml:"tests"`
}

func LoadConfig(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read file %s, error: %v", path, err)
	}
	c := &Config{}
	if err = yaml.Unmarshal(data, c); err != nil {
		log.Fatalf("cannot parse yaml file: %v", err)
	}
	return c
}
