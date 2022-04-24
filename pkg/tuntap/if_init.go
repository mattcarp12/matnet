package tuntap

// Create a new TAP interface with the given name and IP address.
func TapInit(name string, ipAddr string) *Interface {
	iface, err := New(Config{
		DeviceType: TAP,
		PlatformSpecificParams: PlatformSpecificParams{
			Name: name,
		},
	})

	if err != nil {
		panic(err)
	}

	IfaceConfig(name, ipAddr)

	return iface
}
