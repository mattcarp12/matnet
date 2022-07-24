package tuntap

func (iface *Interface) Read(b []byte) (int, error) {
	return iface.ReadWriteCloser.Read(b)
}

func (iface *Interface) Write(b []byte) (int, error) {
	return iface.ReadWriteCloser.Write(b)
}
