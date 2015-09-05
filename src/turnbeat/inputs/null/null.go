package null

/* an empty input, useful as a starting point for a new input */

import (
  "time"
  "libbeat/common"
  "libbeat/logp"
  "turnbeat/inputs"
)

type NullInput struct {
  Config  inputs.MothershipConfig
  Type	  string
}

func (l *NullInput) InputType() string {
  return "NullInput"
}

func (l *NullInput) InputVersion() string {
  return "0.0.1"
}

func (l *NullInput) Init(config inputs.MothershipConfig) error {
  l.Type = "null"
  l.Config = config
  logp.Info("[NullInput] Initialized")
  return nil
}

func (l *NullInput) GetConfig() inputs.MothershipConfig {
  return l.Config
}

// below is an example of a run for an "interrupt" style input
// see bottom for a "periodic" style input
func (l *NullInput) Run(output chan common.MapStr) error {
  logp.Debug("[NullInput]", "Running Null Input")

  // dispatch thread here
  go func(output chan common.MapStr) {
    l.doStuff(output)
  }(output)

  return nil
}

func (l *NullInput) doStuff(output chan common.MapStr) {
  now := func() time.Time {
    t := time.Now()
    return t
  }

  // construct event and write it to channel
  event := common.MapStr{}

  text := "null event"
  event["message"] = &text
  event["type"] = l.Type

  event.EnsureTimestampField(now)
  event.EnsureCountField()

  output <- event


}

// If you had a periodic type input, use the below as the "Run" method instead of the above "Run"
func (l *NullInput) RunPeriodic (output chan common.MapStr) error {
  logp.Debug("[nullinput]", "Starting up Null Input")

  // use the runTick for tick interval, empty functions for minor and major
  go inputs.PeriodicTaskRunner (l, output, l.doStuff, inputs.EmptyFunc, inputs.EmptyFunc)

  return nil
}

