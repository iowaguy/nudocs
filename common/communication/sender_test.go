package communication

import (
	"fmt"
	"testing"

	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/clock"
)

func TestSendToPeers(t *testing.T) {
	clock.NewLocalVectorClock(5, 0)
	po := common.NewPeerOperation("i", "s", 123)

	fmt.Println(po)
}
