package netstack

// function to calculate the checksum of the IPv4 header
func Checksum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)

	// make length multiple of 2
	if length%2 != 0 {
		data = append(data, 0)
		length++
	}

	// sum the 16-bit words
	for index < length {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
	}

	// add top 16 bits to bottom 16 bits
	sum = (sum >> 16) + (sum & 0xffff)

	// return 1's complement of sum
	return uint16(^sum)
}
