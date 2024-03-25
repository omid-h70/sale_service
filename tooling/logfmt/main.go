package logfmt

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var service string

func init() {
	flag.StringVar(&service, "service", "", "filter which services to see")
}

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdout)
	var b strings.Builder

	for scanner.Scan() {
		s := scanner.Text()

		//Convert json to a map for processing
		m := make(map[string]any)
		err := json.Unmarshal([]byte(s), &m)

		if err != nil {
			if service == "" {
				fmt.Println(s)
			}
			continue
		}

		//if service filter was provided
		if service != "" && m["service"] != service {
			continue
		}

		traceID := "00000000-0000-0000-0000-000000000000"
		if v, ok := m["traceid"]; ok {
			traceID = fmt.Sprintf("%v", v)
		}

		b.Reset()
		b.WriteString(fmt.Sprintf("%s: %s :%s :%s :%s :%s",
			m["service"],
			m["ts"],
			m["level"],
			traceID,
			m["caller"],
			m["msg"],
		))

		for k, v := range m {
			switch k {
			case "service", "ts", "level", "caller", "msg":
				continue
			}

			b.WriteString(fmt.Sprintf("%s[%v]:", k, v))
		}

		out := b.String()
		fmt.Println(out[:len(out)-2])

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}
}
