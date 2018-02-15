package allochandlers

import (
	"github.com/adragoset/shepard/watchers/file"
)

//GUIDPATTERN regular expression to match guids
const GUIDPATTERN = "([A-Z0-9a-z]{8}-([A-Z0-9a-z]{4}-){3}[A-Z0-9a-z]{12}){1}"

//FileWatchHandler Interfacce for handling file events
type FileWatchHandler interface {
	HandleEvents(event filewatcher.Event)
}
