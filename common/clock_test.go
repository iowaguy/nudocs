package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVectorClocks(t *testing.T) {
	// five peers
	vc := NewLocalVectorClock(5, 0)
	assert.Equal(t, "[0 0 0 0 0]", vc.String())

	vc1 := NewVectorClock([]int{0, 0, 0, 0, 0})
	assert.Equal(t, vc, vc1)

	vc2 := NewVectorClock([]int{0, 0, 0, 0, 1})
	assert.NotEqual(t, vc, vc2)

	vci := NewVectorClock([]int{1, 0, 0, 0, 0})
	assert.NotEqual(t, vc, vc2)

	assert.False(t, vci.HappenedBefore(vc2))
	assert.False(t, vci.HappenedAfter(vc2))
	assert.True(t, vci.Independent(vc2))

	assert.True(t, vc.HappenedBefore(vc2))
	assert.True(t, vc2.HappenedAfter(vc))

	vc.IncrementClock()
	assert.Equal(t, NewVectorClock([]int{1, 0, 0, 0, 0}), vc)

	vc.IncrementClock()
	assert.Equal(t, NewVectorClock([]int{2, 0, 0, 0, 0}), vc)

	vc.UpdateClock(vc2)
	assert.Equal(t, NewVectorClock([]int{2, 0, 0, 0, 1}), vc)
}
