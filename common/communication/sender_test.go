package communication

import (
	"fmt"
	"testing"
)

func TestSendToPeers(t *testing.T) {
	NewLocalVectorClock(5, 0)
	po := NewPeerOperation("i", "s", 123)

	fmt.Println(po)
}
