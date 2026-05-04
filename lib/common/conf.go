package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// UpdateConfFile rewrites the given key/value pairs in a beego-style
// ini configuration file while preserving comments, ordering and the
// rest of the file. Keys not present in the file are appended at the
// end. The file is rewritten atomically.
func UpdateConfFile(path string, kv map[string]string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	_ = f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	seen := make(map[string]bool, len(kv))
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") || strings.HasPrefix(trim, ";") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if v, ok := kv[key]; ok {
			lines[i] = fmt.Sprintf("%s=%s", key, v)
			seen[key] = true
		}
	}
	for k, v := range kv {
		if !seen[k] {
			lines = append(lines, fmt.Sprintf("%s=%s", k, v))
		}
	}

	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	tmp := path + ".tmp"
	if err := ioutil.WriteFile(tmp, []byte(out), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
