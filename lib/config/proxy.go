// Copyright 2020 Ipalfish, Inc.
// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Workdir  string      `yaml:"workdir,omitempty" toml:"workdir,omitempty" json:"workdir,omitempty"`
	Proxy    ProxyServer `yaml:"proxy,omitempty" toml:"proxy,omitempty" json:"proxy,omitempty"`
	API      API         `yaml:"api,omitempty" toml:"api,omitempty" json:"api,omitempty"`
	Metrics  Metrics     `yaml:"metrics,omitempty" toml:"metrics,omitempty" json:"metrics,omitempty"`
	Log      Log         `yaml:"log,omitempty" toml:"log,omitempty" json:"log,omitempty"`
	Security Security    `yaml:"security,omitempty" toml:"security,omitempty" json:"security,omitempty"`
	Advance  Advance     `yaml:"advance,omitempty" toml:"advance,omitempty" json:"advance,omitempty"`
}

type Metrics struct {
	MetricsAddr     string `toml:"metrics-addr" json:"metrics-addr"`
	MetricsInterval uint   `toml:"metrics-interval" json:"metrics-interval"`
}

type ProxyServerOnline struct {
	MaxConnections uint64 `yaml:"max-connections,omitempty" toml:"max-connections,omitempty" json:"max-connections,omitempty"`
	TCPKeepAlive   bool   `yaml:"tcp-keep-alive,omitempty" toml:"tcp-keep-alive,omitempty" json:"tcp-keep-alive,omitempty"`
}

type ProxyServer struct {
	ProxyServerOnline
	Addr          string `yaml:"addr,omitempty" toml:"addr,omitempty" json:"addr,omitempty"`
	PDAddrs       string `yaml:"pd-addrs,omitempty" toml:"pd-addrs,omitempty" json:"pd-addrs,omitempty"`
	ProxyProtocol string `yaml:"proxy-protocol,omitempty" toml:"proxy-protocol,omitempty" json:"proxy-protocol,omitempty"`
}

type API struct {
	Addr            string `yaml:"addr,omitempty" toml:"addr,omitempty" json:"addr,omitempty"`
	EnableBasicAuth bool   `yaml:"enable-basic-auth,omitempty" toml:"enable-basic-auth,omitempty" json:"enable-basic-auth,omitempty"`
	User            string `yaml:"user,omitempty" toml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty"`
}

type Advance struct {
	PeerPort             string `yaml:"peer-port,omitempty" toml:"peer-port,omitempty" json:"peer-port,omitempty"`
	IgnoreWrongNamespace bool   `yaml:"ignore-wrong-namespace,omitempty" toml:"ignore-wrong-namespace,omitempty" json:"ignore-wrong-namespace,omitempty"`
	WatchInterval        string `yaml:"watch-interval,omitempty" toml:"watch-interval,omitempty" json:"watch-interval,omitempty"`
}

type Log struct {
	Level   string  `yaml:"level,omitempty" toml:"level,omitempty" json:"level,omitempty"`
	Encoder string  `yaml:"encoder,omitempty" toml:"encoder,omitempty" json:"encoder,omitempty"`
	LogFile LogFile `yaml:"log-file,omitempty" toml:"log-file,omitempty" json:"log-file,omitempty"`
}

type LogFile struct {
	Filename   string `yaml:"filename,omitempty" toml:"filename,omitempty" json:"filename,omitempty"`
	MaxSize    int    `yaml:"max-size,omitempty" toml:"max-size,omitempty" json:"max-size,omitempty"`
	MaxDays    int    `yaml:"max-days,omitempty" toml:"max-days,omitempty" json:"max-days,omitempty"`
	MaxBackups int    `yaml:"max-backups,omitempty" toml:"max-backups,omitempty" json:"max-backups,omitempty"`
}

type TLSCert struct {
	CA        string `yaml:"ca,omitempty" toml:"ca,omitempty" json:"ca,omitempty"`
	SkipCA    bool   `yaml:"skip-ca,omitempty" toml:"skip-ca,omitempty" json:"skip-ca,omitempty"`
	Cert      string `yaml:"cert,omitempty" toml:"cert,omitempty" json:"cert,omitempty"`
	Key       string `yaml:"key,omitempty" toml:"key,omitempty" json:"key,omitempty"`
	AutoCerts bool   `yaml:"auto-certs,omitempty" toml:"auto-certs,omitempty" json:"auto-certs,omitempty"`
}

func (c TLSCert) HasCert() bool {
	return !(c.Cert == "" && c.Key == "")
}

func (c TLSCert) HasCA() bool {
	return c.CA != ""
}

type Security struct {
	RSAKeySize int     `yaml:"rsa-key-size,omitempty" toml:"rsa-key-size,omitempty" json:"rsa-key-size,omitempty"`
	Client     TLSCert `yaml:"client,omitempty" toml:"client,omitempty" json:"client,omitempty"`
	Cluster    TLSCert `yaml:"cluster,omitempty" toml:"cluster,omitempty" json:"cluster,omitempty"`
	PDTLS      TLSCert `yaml:"pd-tls,omitempty" toml:"pd-tls,omitempty" json:"pd-tls,omitempty"`
	TiDBTLS    TLSCert `yaml:"tidb-tls,omitempty" toml:"tidb-tls,omitempty" json:"tidb-tls,omitempty"`
}

func NewConfig(data []byte) (*Config, error) {
	var cfg Config
	cfg.Advance.IgnoreWrongNamespace = true
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) Check() error {
	if cfg.Workdir == "" {
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg.Workdir = filepath.Clean(d)
	}
	return nil
}

func (cfg *Config) ToBytes() ([]byte, error) {
	return yaml.Marshal(cfg)
}
