package tuntap

func (iface *Interface) Read(b []byte) (n int, err error) {
	return iface.ReadWriteCloser.Read(b)
}

func (iface *Interface) Write(b []byte) (n int, err error) {
	return iface.ReadWriteCloser.Write(b)
}
