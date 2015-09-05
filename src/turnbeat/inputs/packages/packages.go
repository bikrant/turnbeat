package packages

/* an empty input, useful as a starting point for a new input */

import (
  "bytes"
  "os/exec"
  "time"
  "strings"
  "libbeat/common"
  "libbeat/logp"
  "turnbeat/inputs"
)

type PackagesInput struct {
  Config    inputs.MothershipConfig
  Type      string
}

type RPMPackage struct {
    Name string
    Version string
    Arch  string
}

func (l *PackagesInput) InputType() string {
  return "PackagesInput"
}

func (l *PackagesInput) InputVersion() string {
  return "0.0.1"
}

func (l *PackagesInput) Init(config inputs.MothershipConfig) error {
  l.Config = config
  l.Type = "Packages"
  logp.Info("[PackagesInput] Initialized")
  return nil
}

func (l *PackagesInput) GetConfig() inputs.MothershipConfig {
  return l.Config
}

func (l *PackagesInput) Run(output chan common.MapStr) error {
  logp.Debug("[PackagesInput]", "Running Packages Input")

  // dispatch thread here
  go inputs.PeriodicTaskRunner (l, output, l.doStuff, inputs.EmptyFunc, inputs.EmptyFunc)

  return nil
}

func (l *PackagesInput) doStuff(output chan common.MapStr) {

  now := func() time.Time {
    t := time.Now()
    return t
  }

  // construct event and write it to channel
  event := common.MapStr{}

  //text := "null event"
  //event["message"] = &text

  event["message"]  = "packages event"
  event["type"]     = l.Type

  event.EnsureTimestampField(now)
  event.EnsureCountField()


  /////////////
  cmd := exec.Command("/bin/rpm", "-qa", "--queryformat", "%{NAME}:::%{VERSION}:::%{ARCH}##")
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()
  if err != nil {
    logp.Info("Error occurred")
    return
  }

  items := strings.Split(out.String(),"##")
  rpmList := make([]RPMPackage, 0)

  for _, line := range items {
    item := strings.Split(line, ":::")
    if len(item) < 3{
      continue
    }

    pkg := RPMPackage {
              Name: item[0],
              Version: item[1],
              Arch: item[2],
            }

    rpmList = append(rpmList, pkg)
  }

  event["packages"] = rpmList
  output <- event

}
