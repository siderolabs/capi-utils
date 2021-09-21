// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package capi manages CAPI installation, provides default client for CAPI CRDs.
package capi

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
)

// Config custom implementation of config reader for clusterctl.
type Config struct {
	config      *viper.Viper
	configPaths []string
}

func newConfig() *Config {
	c := viper.New()

	replacer := strings.NewReplacer("-", "_")

	c.SetEnvKeyReplacer(replacer)
	c.AllowEmptyEnv(true)
	c.AutomaticEnv()

	return &Config{
		config:      c,
		configPaths: []string{filepath.Join(homedir.HomeDir(), config.ConfigFolder)},
	}
}

// Init initializes the config.
func (c *Config) Init(path string) error {
	if path != "" {
		url, err := url.Parse(path)
		if err != nil {
			return fmt.Errorf("failed to url parse the config path %w", err)
		}

		switch {
		case url.Scheme == "https" || url.Scheme == "http":
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url.String(), nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(request)
			if err != nil {
				return fmt.Errorf("failed to download the clusterctl config file from %s %w", url, err)
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to download the clusterctl config file from %s got %d", url, resp.StatusCode)
			}

			defer io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck
			defer resp.Body.Close()                  //nolint:errcheck

			if err = c.config.ReadConfig(resp.Body); err != nil {
				return err
			}
		default:
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("failed to check if clusterctl config file exists %w", err)
			}

			c.config.SetConfigFile(path)
		}
	} else {
		// Checks if there is a default .cluster-api/clusterctl{.extension} file in home directory
		if !c.checkDefaultConfig() {
			return nil
		}

		// Configure viper for reading .cluster-api/clusterctl{.extension} in home directory
		c.config.SetConfigName(config.ConfigName)
		for _, p := range c.configPaths {
			c.config.AddConfigPath(p)
		}
	}

	return c.config.ReadInConfig()
}

// Get implements config.Reader.
func (c *Config) Get(key string) (string, error) {
	if c.config.Get(key) == nil {
		return "", fmt.Errorf("failed to get value for variable %q. Please set the variable value using os env variables or using the .clusterctl config file", key)
	}

	return c.config.GetString(key), nil
}

// Set implements config.Reader.
func (c *Config) Set(key, value string) {
	c.config.Set(key, value)
}

// UnmarshalKey implements config.Reader.
func (c *Config) UnmarshalKey(key string, rawval interface{}) error {
	return c.config.UnmarshalKey(key, rawval)
}

// checkDefaultConfig checks the existence of the default config.
// Returns true if it finds a supported config file in the available config
// folders.
func (c *Config) checkDefaultConfig() bool {
	for _, path := range c.configPaths {
		for _, ext := range viper.SupportedExts {
			f := filepath.Join(path, fmt.Sprintf("%s.%s", config.ConfigName, ext))

			_, err := os.Stat(f)
			if err == nil {
				return true
			}
		}
	}

	return false
}
