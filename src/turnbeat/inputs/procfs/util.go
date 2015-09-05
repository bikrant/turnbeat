package procfs

import (
	"strings"
	"bytes"
        "os"
        "regexp"
        "syscall"
)

func pathExists(pathname string) bool {
        _, err := os.Stat(pathname)
        return err != syscall.ENOENT
}

func isNumeric(s string) bool {
        a, _ := regexp.MatchString("^[0-9]+$", s)
        return a
}

func byteTransform(buff []byte) string {
  // replace null bytes with spaces
  b := bytes.Replace(buff[:], []byte{0}, []byte(" "), -1)
  str := string(b[:])
  // trim trailing spaces
  str = strings.Trim(str, " ")
  return str
}
