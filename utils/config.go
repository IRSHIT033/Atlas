package utils

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type LoadBalanceStrategy int

const (
	RoundRobin LoadBalanceStrategy = iota
	LeastConnected
)

func GetLoadBalanceStrategy(strategy string) LoadBalanceStrategy {
	switch strategy {
	case "least-connection":
		return LeastConnected
	default:
		return RoundRobin
	}

}

type Backend struct {
	Url    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type Config struct {
	Port            int       `yaml:"loadbalance_port"`
	MaxAttemptLimit int       `yaml:"max_attempt_limit"`
	Backends        []Backend `yaml:"backends"`
	Strategy        string    `yaml:"strategy"`
}

const MAX_LB_ATTEMPTS int = 3

func GetLBConfig() (*Config, error) {
	var config Config
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	if len(config.Backends) == 0 {
		return nil, errors.New("backend hosts expected, none provided")
	}

	if config.Port == 0 {
		return nil, errors.New("load balancer port not found")
	}

	return &config, nil
}
