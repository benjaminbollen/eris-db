// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package config

import (
  "fmt"

	"github.com/spf13/viper"
	"github.com/naoina/toml"

	"github.com/eris-ltd/eris-db/files"
)

// Standard configuration file for the server.
type (
	ErisDBConfig struct {
		DB         DB           `toml:"db"`
		TMSP       TMSP         `toml:"tmsp"`
		Tendermint Tendermint   `toml:"tendermint"`
		Server     ServerConfig `toml":server"`
	}

	DB struct {
		Backend string `toml:"backend"`
	}

	TMSP struct {
		Listener string `toml:"listener"`
	}

	Tendermint struct {
		Host string `toml:"host"`
	}

	ServerConfig struct {
		Bind      Bind      `toml:"bind"`
		TLS       TLS       `toml:"TLS"`
		CORS      CORS      `toml:"CORS"`
		HTTP      HTTP      `toml:"HTTP"`
		WebSocket WebSocket `toml:"web_socket"`
		Logging   Logging   `toml:"logging"`
	}

	Bind struct {
		Address string `toml:"address"`
		Port    uint16 `toml:"port"`
	}

	TLS struct {
		TLS      bool   `toml:"tls"`
		CertPath string `toml:"cert_path"`
		KeyPath  string `toml:"key_path"`
	}

	// Options stores configurations
	CORS struct {
		Enable           bool     `toml:"enable"`
		AllowOrigins     []string `toml:"allow_origins"`
		AllowCredentials bool     `toml:"allow_credentials"`
		AllowMethods     []string `toml:"allow_methods"`
		AllowHeaders     []string `toml:"allow_headers"`
		ExposeHeaders    []string `toml:"expose_headers"`
		MaxAge           uint64   `toml:"max_age"`
	}

	HTTP struct {
		JsonRpcEndpoint string `toml:"json_rpc_endpoint"`
	}

	WebSocket struct {
		WebSocketEndpoint    string `toml:"websocket_endpoint"`
		MaxWebSocketSessions uint   `toml:"max_websocket_sessions"`
		ReadBufferSize       uint   `toml:"read_buffer_size"`
		WriteBufferSize      uint   `toml:"write_buffer_size"`
	}

	Logging struct {
		ConsoleLogLevel string `toml:"console_log_level"`
		FileLogLevel    string `toml:"file_log_level"`
		LogFile         string `toml:"log_file"`
		VMLog           bool   `toml:"vm_log"`
	}
)

// Initialise

func DefaultServerConfig() ServerConfig {
	cp := ""
	kp := ""
	return ServerConfig{
		Bind: Bind{
			Address: "",
			Port:    1337,
		},
		TLS: TLS{TLS: false,
			CertPath: cp,
			KeyPath:  kp,
		},
		CORS: CORS{},
		HTTP: HTTP{JsonRpcEndpoint: "/rpc"},
		WebSocket: WebSocket{
			WebSocketEndpoint:    "/socketrpc",
			MaxWebSocketSessions: 50,
			ReadBufferSize:       4096,
			WriteBufferSize:      4096,
		},
		Logging: Logging{
			ConsoleLogLevel: "info",
			FileLogLevel:    "warn",
			LogFile:         "",
		},
	}
}

func DefaultErisDBConfig() *ErisDBConfig {
	return &ErisDBConfig{
		DB: DB{
			Backend: "leveldb",
		},
		TMSP: TMSP{
			Listener: "tcp://0.0.0.0:46658",
		},
		Tendermint: Tendermint{
			Host: "0.0.0.0:46657",
		},
		Server: DefaultServerConfig(),
	}
}

// Read a TOML server configuration file.
func ReadErisDBConfig(filePath string) (*ErisDBConfig, error) {
	bts, err := files.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	cfg := &ErisDBConfig{}
	err2 := toml.Unmarshal(bts, cfg)
	if err2 != nil {
		return nil, err2
	}
	return cfg, nil
}

// Write a server configuration file.
func WriteErisDBConfig(filePath string, cfg *ErisDBConfig) error {
	bts, err := toml.Marshal(*cfg)
	if err != nil {
		return err
	}
	return files.WriteAndBackup(filePath, bts)
}