package tcp

import (
  "net"
  "bufio"
  "bytes"
  "strconv"
  "io"
  "regexp"
  "time"
  "strings"
  "errors"
  "github.com/johann8384/libbeat/common"
  "github.com/johann8384/libbeat/logp"
  "github.com/turn/turnbeat/inputs"
)

//type TSDBMetricExp struct {
//  *regexp.Regexp
//}

//var metricExp = TSDBMetricExp{regexp.MustCompile(`^put (?P<metric_name>[\w.]+)[\s]+(?P<metric_timestamp>[0-9]+)[\s]+(?P<metric_value>[0-9.]+)[\s]+(?P<metric_tags>.*$)`)}

//func (r *TSDBMetricExp) FindStringSubmatchMap(s string) map[string]string {
//  captures := make(map[string]string)
//
//  match := r.FindStringSubmatch(s)
//  if match == nil {
//    return captures
//  }
//
//  for i, name := range r.SubexpNames() {
//    if i == 0 {
//      continue
//    }
//    captures[name] = match[i]
//
//  }
//  return captures
//}

type TcpInput struct {
  Port       int /* the port to listen on */
  Type       string /* the type to add to events */
}

func (l *TcpInput) InputType() string {
  return "TcpInput"
}

func (l *TcpInput) InputVersion() string {
  return "0.0.1"
}

func (l *TcpInput) Init(config inputs.MothershipConfig) error {

  if config.Port == 0 {
    return errors.New("No Input Port specified")
  }
  l.Port = config.Port

  if config.Type == "" {
    return errors.New("No Event Type specified")
  }
  l.Type = config.Type

  logp.Info("[TcpInput] Using Port %d", l.Port)
  logp.Info("[TcpInput] Adding Event Type %s", l.Type)

  return nil
}

func (l *TcpInput) Run(output chan common.MapStr) error {
  logp.Debug("tcpinput", "Running TCP Input")
  server, err := net.Listen("tcp", ":" + strconv.Itoa(l.Port))
  if server == nil {
      panic("couldn't start listening: " + err.Error())
  }
  conns := clientConns(server)
  for {
    go l.handleConn(<-conns, output)
  }
}

func clientConns(listener net.Listener) chan net.Conn {
    ch := make(chan net.Conn)
    i := 0
    go func() {
        for {
            client, err := listener.Accept()
            if client == nil {
                logp.Info("couldn't accept: " + err.Error())
                continue
            }
            i++
            logp.Debug("tcpinput", "%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
            ch <- client
        }
    }()
    return ch
}

func (l *TcpInput) handleConn(client net.Conn, output chan common.MapStr) {
    reader := bufio.NewReader(client)
    buffer := new(bytes.Buffer)

    var source string = client.RemoteAddr().String()
    var offset int64 = 0
    var line uint64 = 0
    var read_timeout = 10 * time.Second

    logp.Debug("tcpinput", "Handling New Connection from %s", source)

    now := func() time.Time {
      t := time.Now()
      return t
    }

    for {
        text, bytesread, err := l.readline(reader, buffer, read_timeout)

        if err != nil {
          logp.Info("Unexpected state reading from %v; error: %s\n", client.RemoteAddr().String, err)
          return
        }

        logp.Debug("tcpinputlines", "New Line: %s", &text)

//        metric_data := metricExp.FindStringSubmatchMap(*text)
//        parsed_tags := strings.Fields(metric_data["metric_tags"])
//        tags := make(map[string]string)

//        for  _,v := range parsed_tags {
//          tag := strings.Split(v, "=")
//          tags[tag[0]] = tag[1]
//        }

        line++

        event := common.MapStr{}
        event["source"] = &source
        event["offset"] = offset
        event["line"] = line
        event["message"] = text
        event["type"] = l.Type
//        event["metric_name"] = metric_data["metric_name"]
//        event["metric_value"] = metric_data["metric_value"]
//        event["metric_timestamp"] = metric_data["metric_timestamp"]
//        event["metric_tags"] = metric_data["metric_tags"]
//        event["metric_tags_map"] = tags

        event.EnsureTimestampField(now)
        event.EnsureCountField()

        offset += int64(bytesread)

        logp.Debug("tcpinput", "InputEvent: %v", event)
        output <- event // ship the new event downstream
        client.Write([]byte("OK"))
    }
    logp.Debug("tcpinput", "Closed Connection from %s", source)
}

func (l *TcpInput) readline(reader *bufio.Reader, buffer *bytes.Buffer, eof_timeout time.Duration) (*string, int, error) {
  var is_partial bool = true
  var newline_length int = 1
  start_time := time.Now()
  
  logp.Debug("tcpinputlines", "Readline Called")

  for {
    segment, err := reader.ReadBytes('\n')

    if segment != nil && len(segment) > 0 {
      if segment[len(segment)-1] == '\n' {
        // Found a complete line
        is_partial = false

        // Check if also a CR present
        if len(segment) > 1 && segment[len(segment)-2] == '\r' {
          newline_length++
        }
      }

      // TODO(sissel): if buffer exceeds a certain length, maybe report an error condition? chop it?
      buffer.Write(segment)
    }

    if err != nil {
      if err == io.EOF && is_partial {
        time.Sleep(1 * time.Second) // TODO(sissel): Implement backoff

        // Give up waiting for data after a certain amount of time.
        // If we time out, return the error (eof)
        if time.Since(start_time) > eof_timeout {
          return nil, 0, err
        }
        continue
      } else {
        //emit("error: Harvester.readLine: %s", err.Error())
        return nil, 0, err // TODO(sissel): don't do this?
      }
    }

    // If we got a full line, return the whole line without the EOL chars (CRLF or LF)
    if !is_partial {
      // Get the str length with the EOL chars (LF or CRLF)
      bufferSize := buffer.Len()
      str := new(string)
      *str = buffer.String()[:bufferSize-newline_length]
      // Reset the buffer for the next line
      buffer.Reset()
      return str, bufferSize, nil
    }
  } /* forever read chunks */

  return nil, 0, nil
}
