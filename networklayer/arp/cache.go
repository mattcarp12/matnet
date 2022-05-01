package arp

import (
	"errors"
	"net"
	"time"
)

// TODO: Start goroutine to clean up cache periodically

type ARPCacheEntry struct {
	MAC       net.HardwareAddr
	timestamp time.Time
}

type ARPCache map[string]ARPCacheEntry

const ARPTimeout = 5

func NewARPCache() *ARPCache {
	return &ARPCache{}
}

func (c *ARPCache) Update(h *ARPHeader) {
	ip := h.SourceIPAddr.String()

	(*c)[ip] = ARPCacheEntry{
		MAC:       h.SourceHWAddr,
		timestamp: time.Now(),
	}
}

func (c *ARPCache) Cleanup() {
	now := time.Now()
	for ip, entry := range *c {
		if now.Sub(entry.timestamp) > ARPTimeout*time.Second {
			delete(*c, ip)
		}
	}
}

var ErrArpCacheMiss error = errors.New("arp cache miss")

func (c *ARPCache) Lookup(ip net.IP) (net.HardwareAddr, error) {
	if entry, ok := (*c)[ip.String()]; ok {
		return entry.MAC, nil
	}

	return nil, ErrArpCacheMiss
}
