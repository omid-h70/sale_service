package docker

import (
	"bytes"
	"encoding/json"
	"net"
	"os/exec"
	"testing"
)

type Container struct {
	ID   string
	Host string
}

func StartContainer(t *testing.T, image string, port string, args ...string) *Container {
	arg := []string{"run", "-P", "-d"}
	arg = append(arg, args...)
	arg = append(arg, image)

	cmd := exec.Command("docker", arg...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start the container  %s %v ", image, err)
	}

	id := out.String()[:12]

	var doc []map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json %v ", err)
	}

	ip, randPort := extractIPPort(t, doc, port)
	c := Container{
		ID:   id,
		Host: net.JoinHostPort(ip, randPort),
	}

	t.Logf("Image %s", image)
	t.Logf("Container ID %s", c.ID)
	t.Logf("Host %s", c.Host)

	return &c
}

func StopContainer(t *testing.T, id string) {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		t.Fatalf("could not stop container %v ", err)
	}

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("could not remove container %v ", err)
	}
	t.Log("Removed:", id)
}

func DumpContainerLogs(t *testing.T, id string) {

	out, err := exec.Command("docker", "rm", id, "-v").CombinedOutput()
	if err != nil {
		t.Fatalf("could not remove container %v ", err)
	}
	t.Logf("Logs For %s \n %s:", id, out)
}

func extractIPPort(t *testing.T, doc []map[string]any, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		t.Fatalf("could not get network settings")
	}

	ports, ok := nw.(map[string]any)["Ports"]
	if !ok {
		t.Fatalf("could not get ports settings")
	}

	tcp, ok := ports.(map[string]any)[port+"/tcp"]
	if !ok {
		t.Fatalf("could not get tcp settings")
	}

	list, ok := tcp.([]any)
	if !ok {
		t.Fatalf("could not get network list tcp settings")
	}

	var hostIp string
	var hostPort string
	for _, l := range list {
		data, ok := l.(map[string]any)
		if !ok {
			t.Fatalf("could not get network ports tcp data")
		}

		hostIp = data["HostIp"].(string)
		if hostIp != "::" {
			hostPort = data["HostPort"].(string)
		}
	}
	return hostPort, hostIp
}
