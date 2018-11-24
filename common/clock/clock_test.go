package clock

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVectorClocks(t *testing.T) {
	vc := NewLocalVectorClock(5, 0)
	vc1 := NewVectorClock([]int{0, 0, 0, 0, 0})
	vc2 := NewVectorClock([]int{0, 0, 0, 0, 1})

	assert.Equal(t, "[0 0 0 0 0]", vc.String())
	assert.Equal(t, vc, vc1)
	assert.NotEqual(t, vc, vc2)
	assert.NotEqual(t, vc, vc2)
}

func TestIncrementVectorClock(t *testing.T) {
	vc := NewLocalVectorClock(5, 0)

	vc.IncrementClock()
	assert.Equal(t, NewVectorClock([]int{1, 0, 0, 0, 0}), vc)

	vc.IncrementClock()
	assert.Equal(t, NewVectorClock([]int{2, 0, 0, 0, 0}), vc)
}

func TestUpdateVectorClock(t *testing.T) {
	vc := NewLocalVectorClock(5, 0)
	vc2 := NewVectorClock([]int{0, 0, 0, 0, 1})

	vc.UpdateClock(vc2)
	assert.Equal(t, NewVectorClock([]int{2, 0, 0, 0, 1}), vc)
}

func TestCausality(t *testing.T) {
	vc := NewLocalVectorClock(5, 0)
	vc2 := NewVectorClock([]int{0, 0, 0, 0, 1})
	vci := NewVectorClock([]int{1, 0, 0, 0, 0})

	assert.False(t, vci.HappenedBefore(vc2))
	assert.False(t, vci.HappenedAfter(vc2))
	assert.True(t, vci.Independent(vc2))

	fmt.Println(vc)
	fmt.Println(vc2)
	assert.True(t, vc2.HappenedBefore(vc))
	assert.True(t, vc.HappenedAfter(vc2))
}

func TestVectorClockParsing(t *testing.T) {
	vcp, err := ParseVectorClock("[0 0 0 0]")
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, NewVectorClock([]int{0, 0, 0, 0}), vcp)

	vcp, err = ParseVectorClock("[0]")
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, NewVectorClock([]int{0}), vcp)

	vcp, err = ParseVectorClock("[0]")
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, NewVectorClock([]int{1}), vcp.IncrementClock())
}
