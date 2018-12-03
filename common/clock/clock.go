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

func NewVectorClock(other []int, pid int) *VectorClock {
	vc := new(VectorClock)
	vc.state = other
	vc.localPid = pid
	return vc
}

func (me *VectorClock) IncrementClock() *VectorClock {
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

func (me *VectorClock) CausallyPreceding(other *VectorClock) bool {
	// causally preceding iff the following 2 conditions are met
	// (1) SVo[s] = SVd[s] + 1
	if (other.state[other.localPid] - me.state[other.localPid]) != 1 {
		return false
	}

	// (2) SVo[i] <= SVd[i], for all i except i != s
	for i := range me.state {
		if other.localPid == i {
			continue
		} else if other.state[i] > me.state[i] {
			return false
		}
	}
	return true
}

func (me *VectorClock) String() string {
	return fmt.Sprintf("%v:%v", me.localPid, me.state)
}

func ParseVectorClock(vc string) (*VectorClock, error) {
	// trim white space
	trimmedVcAndPid := strings.TrimSpace(vc)

	vcAndPid := strings.Split(trimmedVcAndPid, ":")
	pid, err := strconv.Atoi(vcAndPid[0])
	if err != nil {
		log.Error("Error: could not parse vector clock pid")
		return &VectorClock{}, err
	}

	trimmedVc := vcAndPid[1][1 : len(vcAndPid[1])-1]
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

	return NewVectorClock(iArr, pid), nil
}
