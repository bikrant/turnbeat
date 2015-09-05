package syslog

import (
  "errors"
  "fmt"
  "time"
  "libbeat/common"
  "libbeat/logp"
  "turnbeat/inputs"
  "gopkg.in/mcuadros/go-syslog.v2"
)

type SyslogInput struct {
  Config     inputs.MothershipConfig
  Port       int /* the port to listen on */
  Type       string /* the type to add to events */
}

func (l *SyslogInput) InputType() string {
  return "SyslogInput"
}

func (l *SyslogInput) InputVersion() string {
  return "0.0.1"
}

func (l *SyslogInput) Init(config inputs.MothershipConfig) error {

  l.Config = config
  if config.Port == 0 {
    return errors.New("No Input Port specified")
  }
  l.Port = config.Port

  if config.Type == "" {
    return errors.New("No Event Type specified")
  }
  l.Type = config.Type

  logp.Info("[SyslogInput] Using Port %d", l.Port)
  logp.Info("[SyslogInput] Adding Event Type %s", l.Type)

  return nil
}

func (l *SyslogInput) GetConfig() inputs.MothershipConfig {
  return l.Config
}

func (l *SyslogInput) Run(output chan common.MapStr) error {
  logp.Debug("sysloginput", "Running Syslog Input")
  logp.Debug("sysloginput", "Listening on %d", l.Port)

  listen := fmt.Sprintf("0.0.0.0:%d", l.Port)

  channel := make(syslog.LogPartsChannel)
  handler := syslog.NewChannelHandler(channel)

  server := syslog.NewServer()
  server.SetFormat(syslog.Automatic)
  server.SetHandler(handler)
  err := server.ListenUDP(listen)
  if err != nil {
      logp.Err("couldn't start ListenUDP: " + err.Error())
  }
  err = server.ListenTCP(listen)
  if err != nil {
      logp.Err("couldn't start ListenTCP: " + err.Error())
  }
  err = server.Boot()
  if err != nil {
      logp.Err("couldn't start server.Boot(): " + err.Error())
  }


  go func(channel syslog.LogPartsChannel, output chan common.MapStr) {
    var line uint64 = 0
     
    now := func() time.Time {
      t := time.Now()
      return t
    }

    for logParts := range channel {
      logp.Debug("sysloginput", "InputEvent: %v", logParts)

      line++
      event := common.MapStr{}
      event["line"] = line
      event["type"] = l.Type

      for k, v := range logParts {
        event[k] = v
      }

      event["source"] = event["client"].(string)

      if event["message"] != nil {
        message := event["message"].(string)
        event["message"] = &message
      } else if event["content"] != nil {
        message := event["content"].(string)
        event["message"] = &message
      }

      // This syslog parser uses the standard name "tag"
      // which is usually the program that wrote it.
      // The logstash syslog_pri puts "program" for this field.
      if event["tag"] != nil {
        program := event["tag"].(string)
        event["program"] = &program
      }

      event.EnsureTimestampField(now)
      event.EnsureCountField()

      logp.Debug("sysloginput", "Output Event: %v", event)
      output <- event // ship the new event downstream
    }
  }(channel,output)

  return nil
}
