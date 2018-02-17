package ink-miner

import (
	"testing"

)

func TestIntersect(t *testing.T) {
	// test stuff here...
	intersect = doesIntersect()
	if intersect != false {
		t.Error("Expected false, got: ", intersect)
	}
}
