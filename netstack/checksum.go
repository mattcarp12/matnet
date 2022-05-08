package netstack

// function to calculate the checksum of the IPv4 header
func Checksum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)

	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}

	if length > 0 {
		sum += uint32(data[index])
	}

	sum += (sum >> 16)

	return uint16(^sum)
}