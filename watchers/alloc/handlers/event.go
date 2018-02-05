package allochandlers

import (
	"github.com/satori/go.uuid"
)

// An Op is a type that is used to describe what type
// of event has occurred during the watching process.
type Op uint32

// Ops
const (
	AllocCreated Op = iota
	AllocRemoved
	SharedAllocConfigChanged
	TaskCreated
	TaskConfigChanged
)

var ops = map[Op]string{
	AllocCreated:             "ALLOCCRT",
	AllocRemoved:             "ALLOCRMV",
	SharedAllocConfigChanged: "SHDALLOCCFGCHG",
	TaskCreated:              "TACRT",
	TaskConfigChanged:        "TACFGCHG",
}

func (e Op) String() string {
	if op, found := ops[e]; found {
		return op
	}
	return "???"
}

//Event Allocation Task
type Event struct {
	Op
	AllocationID uuid.UUID
	TaskID       uuid.UUID
}
