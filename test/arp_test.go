package test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/stretchr/testify/assert"
)

func TestARPReply(t *testing.T) {
	// Make arp request
	out, err := exec.Command("arping", "-c", "1", netstack.DefaultIPAddr).Output()
	assert.NoError(t, err)

	// Get second line of output
	lines := strings.Split(string(out), "\n")
	line := lines[1]

	// Remove response time from string
	line = line[:strings.LastIndex(line, " ")-1]

	// Check if line matches expected output
	expected := "Unicast reply from " + netstack.DefaultIPAddr + " [" + netstack.DefaultMACAddr + "]"
	assert.Equal(t, expected, line)
}
