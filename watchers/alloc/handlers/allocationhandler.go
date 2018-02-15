package allochandlers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adragoset/shepard/watchers/file"
	"github.com/satori/go.uuid"
)

//ALLOCPATHPATTERN regular expression to match allocation directories
const ALLOCPATHPATTERN = "^(%s){1}%s$"
const ALLOCATIONPATHBLACKLISTEXP = "^(%s){1}((?!%s.+)$"

//AllocationHandler manages handling events relating to allocation creation or removal
type AllocationHandler struct {
	FolderPath        string
	EventChannel      chan Event
	ErrorChannel      chan error
	FilterExpressions []string
	pathMatcher       *regexp.Regexp
}

//NewAllocationHandler creates a handler for handling allocation level events
func NewAllocationHandler(folderPath string, eventChannel chan Event, errorChannel chan error) (*AllocationHandler, error) {
	var allocationHandler *AllocationHandler

	allocationHandler = &AllocationHandler{EventChannel: eventChannel, ErrorChannel: errorChannel, FilterExpressions: make([]string, 0)}
	allocationHandler.setFolderPath(folderPath)
	err := allocationHandler.buildRegularExpressions()
	if err != nil {
		err = fmt.Errorf("AllocationHandler failed to compile allocation folder matcher: %s", err.Error())
	}

	return allocationHandler, err
}

//HandleEvents for top level allocations
func (ah *AllocationHandler) HandleEvents(event filewatcher.Event) {
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

func (ah *AllocationHandler) setFolderPath(folderPath string) {
	folderRegex := strings.Replace(folderPath, "\\", "\\\\", -1)
	folderRegex = strings.Replace(folderRegex, "/", "\\\\", -1)

	ah.FolderPath = folderRegex
}

func (ah *AllocationHandler) buildRegularExpressions() error {
	folderRegex := fmt.Sprintf(ALLOCPATHPATTERN, ah.FolderPath, GUIDPATTERN)
	rg, err := regexp.Compile(folderRegex)
	if err != nil {
		return fmt.Errorf("AllocationHandler failed to compile allocation folder matcher: %s", err.Error())
	}
	ah.pathMatcher = rg
	ah.buildFilterExpressions()
	return nil
}

func (ah *AllocationHandler) buildFilterExpressions() {
	blacklistRegex := fmt.Sprintf(ALLOCATIONPATHBLACKLISTEXP, ah.FolderPath, GUIDPATTERN)
	ah.FilterExpressions = append(ah.FilterExpressions, blacklistRegex)
}

func (ah AllocationHandler) handlesEvent(event filewatcher.Event) Event {
	var result Event

	if event.IsDir() {
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
	case filewatcher.Create:
		e = Event{AllocationID: allocID, Op: AllocCreated}
	case filewatcher.Remove:
		e = Event{AllocationID: allocID, Op: AllocRemoved}
	}

	return e
}
