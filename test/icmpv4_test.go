package test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/stretchr/testify/assert"
)

func TestICMPv4Echo(t *testing.T) {
	out, err := exec.Command("ping", "-c", "1", "-4", netstack.DefaultIPAddr).Output()
	assert.NoError(t, err)

	// Get second line of output
	line := strings.Split(string(out), "\n")[1]

	// Get first part of line
	line = strings.Split(line, ":")[0]

	// Check line is as expected
	expected := "64 bytes from " + netstack.DefaultIPAddr

	assert.Equal(t, expected, line)
}
