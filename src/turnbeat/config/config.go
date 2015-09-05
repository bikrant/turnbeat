package config

import (
	"libbeat/common/droppriv"
  "libbeat/logp"
	"libbeat/outputs"
	"turnbeat/inputs"
	"libbeat/publisher"
)

type Config struct {
	Input      map[string]inputs.MothershipConfig
	Output     map[string]outputs.MothershipConfig
	Shipper    publisher.ShipperConfig
	RunOptions droppriv.RunOptions
  Logging    logp.Logging
	Filter     map[string]interface{}
}

// Config Singleton
var ConfigSingleton Config
