package transportlayer

import (
	"testing"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/stretchr/testify/assert"
)

var data = []byte("Hello World")

var dataSize = len(data)

func genTCPSkb(seqNum uint32) TCPBuffer {
	data := data
	skb := netstack.NewSkBuff(data)

	tcpHeader := TCPHeader{
		SeqNum: seqNum,
	}

	return TCPBuffer{
		SkBuff: skb,
		Header: tcpHeader,
	}
}

// ===========================================================================
// Test TCP Rx Queue
// ===========================================================================

func Test_TCP_RxQueue(t *testing.T) {
	// Create a new TCP protocol object
	tcp := NewTCP()

	// Create a new TCB object
	connID := "conn1"
	tcb := tcp.NewTCB(connID)

	t.Logf("TCB: %+v\n", tcb)

	skb1 := genTCPSkb(0)
	skb2 := genTCPSkb(uint32(dataSize))
	skb3 := genTCPSkb(uint32(dataSize * 2))

	t.Logf("skb1: %+v\n", skb1)

	// Add the skbs to the queue (in unsorted order)
	tcb.sortSegment(skb3)
	tcb.sortSegment(skb2)
	tcb.sortSegment(skb1)

	t.Logf("TCB: %+v\n", tcb)
	t.Logf("RxQueue: %+v\n", tcb.RxQueue)

	// Check the skbs are in the correct order
	var skb *netstack.SkBuff

	if len(tcb.RxChanSorted) > 0 {
		skb = <-tcb.RxChanSorted
		t.Logf("SKB: %+v\n", skb)
		assert.Equal(t, skb1.SkBuff, skb)
	} else {
		t.Errorf("RxChanSorted is empty")
	}

	if len(tcb.RxChanSorted) > 0 {
		skb = <-tcb.RxChanSorted
		t.Logf("SKB: %+v\n", skb)
		assert.Equal(t, skb2.SkBuff, skb)
	} else {
		t.Errorf("RxChanSorted is empty")
	}

	if len(tcb.RxChanSorted) > 0 {
		skb = <-tcb.RxChanSorted
		t.Logf("SKB: %+v\n", skb)
		assert.Equal(t, skb3.SkBuff, skb)
	} else {
		t.Errorf("RxChanSorted is empty")
	}
}
