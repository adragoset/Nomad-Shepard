package shepard

import (
	"NomadShepard/nomadclient"
	"net/url"

	"github.com/radovskyb/watcher"
)

//Config struct
type Config struct {
}

//Shepard struct
type Shepard struct {
	NomadURL    *url.URL
	nClient     *nomadclient.NomadServer
	fileWatcher *watcher.Watcher
}

// New Shepard
func New() *Shepard {
	fileWatcher := watcher.New()
	nomadURL, err := url.Parse("")

	return &Shepard{NomadURL: nomadUrl, fileWatcher: fileWatcher}
}
