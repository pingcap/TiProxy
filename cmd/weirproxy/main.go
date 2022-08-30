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

package main

import (
	"io/ioutil"

	"github.com/pingcap/TiProxy/pkg/config"
	"github.com/pingcap/TiProxy/pkg/server"
	"github.com/pingcap/TiProxy/pkg/util/cmd"
	"github.com/pingcap/TiProxy/pkg/util/waitgroup"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "weirproxy",
		Short: "start the proxy server",
	}

	configFile := rootCmd.PersistentFlags().String("config", "conf/weirproxy.yaml", "weir proxy config file path")
	logEncoder := rootCmd.PersistentFlags().String("log_encoder", "", "log in format of tidb, newtidb, console, or json")
	logLevel := rootCmd.PersistentFlags().String("log_level", "", "log level")
	namespaceFiles := rootCmd.PersistentFlags().String("namespaces", "", "import namespace from dir")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		proxyConfigData, err := ioutil.ReadFile(*configFile)
		if err != nil {
			return err
		}

		cfg, err := config.NewConfig(proxyConfigData)
		if err != nil {
			return err
		}

		if *logEncoder != "" {
			cfg.Log.Encoder = *logEncoder
		}
		if *logLevel != "" {
			cfg.Log.Level = *logLevel
		}

		zapcfg := zap.NewDevelopmentConfig()
		zapcfg.Encoding = cfg.Log.Encoder
		if level, err := zap.ParseAtomicLevel(cfg.Log.Level); err == nil {
			zapcfg.Level = level
		}
		logger, err := zapcfg.Build()
		if err != nil {
			return err
		}
		logger = logger.Named("main")

		srv, err := server.NewServer(cmd.Context(), cfg, logger, *namespaceFiles)
		if err != nil {
			logger.Error("fail to create server", zap.Error(err))
			return err
		}

		var wg waitgroup.WaitGroup
		wg.Run(func() {
			if err := srv.Run(cmd.Context()); err != nil {
				logger.Error("shutdown with error", zap.Error(err))
			}
		})

		<-cmd.Context().Done()
		if e := srv.Close(); e != nil {
			logger.Error("shutdown with errors", zap.Error(e))
		} else {
			logger.Info("gracefully shutdown")
		}

		wg.Wait()
		return nil
	}

	cmd.RunRootCommand(rootCmd)
}
