package allochandlers

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	filewatcher "github.com/radovskyb/watcher"
	"github.com/satori/go.uuid"
)

// An Op is a type that is used to describe what type
// of event has occurred during the watching process.
type Op uint32

//PATHRGX regular expression to match allocation directories
const PATHRGX = "^(%s){1}([A-Z0-9]{8}-([A-Z0-9]{4}-){3}[A-Z0-9]{12})$"

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

//AllocationHandler manages handling events relating to allocation creation or removal
type AllocationHandler struct {
	FolderPath   string
	EventChannel chan Event
	ErrorChannel chan error
	pathMatcher  *regexp.Regexp
}

//NewAllocationHandler creates a handler for handling allocation level events
func NewAllocationHandler(folderPath string, eventChannel chan Event, errorChannel chan error) (*AllocationHandler, error) {
	var allocationHandler *AllocationHandler
	folderRegex := strings.Replace(folderPath, "\\", "\\\\", -1)
	folderRegex = strings.Replace(folderRegex, "/", "\\\\", -1)
	folderRegex = fmt.Sprintf(PATHRGX, folderRegex)
	log.Println(folderRegex)
	rg, err := regexp.Compile(folderRegex)
	if err != nil {
		err = fmt.Errorf("AllocationHandler failed to compile allocation folder matcher: %s", err.Error())
	}

	allocationHandler = &AllocationHandler{FolderPath: folderPath, EventChannel: eventChannel, ErrorChannel: errorChannel, pathMatcher: rg}

	return allocationHandler, err
}

//HandleAllocationEvents for top level allocations
func (ah AllocationHandler) HandleAllocationEvents(event filewatcher.Event) {
	allocEvent := ah.handlesEvent(event)
	if allocEvent != (Event{}) {
		go func(e Event) {
			select {
			case ah.EventChannel <- e:
			default:
			}
		}(allocEvent)
	}
}

func (ah AllocationHandler) handlesEvent(event filewatcher.Event) Event {
	var result Event

	if event.IsDir() {
		log.Println(event.Path)
		if ah.pathMatcher.MatchString(event.Path) {
			result = ah.createEvent(event)
		}
	}

	return result
}

func (ah AllocationHandler) createEvent(fileEvent filewatcher.Event) Event {
	var e Event
	guidString := fileEvent.Name()
	allocID, err := uuid.FromString(guidString)
	if err != nil {
		ah.ErrorChannel <- fmt.Errorf("AllocationHandler failed getting AllocationId from FileEvent:%s Error:%s", fileEvent, err.Error())
		return e
	}

	switch fileEventType := fileEvent.Op; fileEventType {
	case filewatcher.Write:
		e = Event{AllocationID: allocID, Op: AllocCreated}
	case filewatcher.Remove:
		e = Event{AllocationID: allocID, Op: AllocRemoved}
	}

	return e
}
