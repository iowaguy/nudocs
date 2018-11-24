package clock

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type VectorClock struct {
	localPid int
	state    []int
}

// a singleton
var clock *VectorClock
var vcOnce sync.Once

func GetLocalVectorClock() *VectorClock {
	return clock
}

// creates the singleton
func NewLocalVectorClock(peers, pid int) *VectorClock {
	vcOnce.Do(func() {
		clock = new(VectorClock)
		clock.state = make([]int, peers)
		clock.localPid = pid
	})

	return clock
}

func NewVectorClock(other []int) *VectorClock {
	vc := new(VectorClock)
	vc.state = other
	return vc
}

func (me *VectorClock) IncrementClock() *VectorClock {
	log.Info("pid=", me.localPid, "; state len=", len(me.state))
	me.state[me.localPid]++
	return me
}

func (me *VectorClock) UpdateClock(other *VectorClock) *VectorClock {
	// make sure length is the same
	if len(me.state) != len(other.state) {
		log.Panic("Error: vector clocks are not the same length: ", len(me.state), " and ", len(other.state))
	}

	for i, v := range me.state {
		if v < other.state[i] {
			me.state[i] = other.state[i]
		}
	}
	return me
}

// true if me happened before other
func (me *VectorClock) HappenedBefore(other *VectorClock) bool {
	for i, v := range me.state {
		if v > other.state[i] {
			return false
		}
	}
	return true
}

func (me *VectorClock) HappenedAfter(other *VectorClock) bool {
	for i, v := range me.state {
		if v < other.state[i] {
			return false
		}
	}
	return true
}

func (me *VectorClock) Independent(other *VectorClock) bool {
	return !me.HappenedBefore(other) && !me.HappenedAfter(other)
}

func (me *VectorClock) String() string {
	return fmt.Sprintf("%v", me.state)
}

func ParseVectorClock(vc string) (*VectorClock, error) {
	// trim white space and brackets
	trimmedVc := strings.TrimSpace(vc)[1 : len(vc)-1]
	vcStringArr := strings.Split(trimmedVc, " ")
	iArr := make([]int, 0, len(vcStringArr))

	for _, v := range vcStringArr {
		val, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Error: could not parse vector clock")
			return &VectorClock{}, err
		}
		iArr = append(iArr, val)
	}

	return NewVectorClock(iArr), nil
}
