package allocwatcher

import (
	"NomadShepard/watchers/alloc/handlers"
	"fmt"
	"time"

	filewatcher "github.com/radovskyb/watcher"
)

//Config struct
type Config struct {
	NomadAllocDir string        `json:"nomad_alloc_dir"`
	WatchCycleMs  time.Duration `json:"watch_cycle_ms"`
}

//Watcher struct
type Watcher struct {
	AllocationEvents  chan allochandlers.Event
	AllocationHandler *allochandlers.AllocationHandler
	ErrorEvents       chan error
	WatchCycleMs      time.Duration
	filewatcher       *filewatcher.Watcher
}

//New Watcher
func New(config *Config) (*Watcher, error) {
	var allocWatcher *Watcher
	fileWatcher := filewatcher.New()
	allocEventChannel := make(chan allochandlers.Event)
	errorEventChannel := make(chan error)
	allocHandler, err := allochandlers.NewAllocationHandler(config.NomadAllocDir, allocEventChannel, errorEventChannel)

	if err != nil {
		err = fmt.Errorf("AllocationWatcher failed to create an AllocationHandler Error:%s", err.Error())
		return allocWatcher, err
	}

	allocWatcher = &Watcher{AllocationEvents: allocEventChannel, AllocationHandler: allocHandler, ErrorEvents: errorEventChannel, WatchCycleMs: config.WatchCycleMs, filewatcher: fileWatcher}
	// Watch the alloc folder for changes.
	err = allocWatcher.Configure(config, fileWatcher)
	if err != nil {
		var emptyWatcher *Watcher
		err = fmt.Errorf("AllocationWatcher failed to create an AllocationHandler Error:%s", err.Error())
		return emptyWatcher, err
	}

	return allocWatcher, nil
}

//Start the Watcher
func (watcher Watcher) Start() {
	watcher.filewatcher.Start(watcher.WatchCycleMs)
}

//Stop the Watcher
func (watcher Watcher) Stop() {
	watcher.filewatcher.Close()
}

//Configure a new Watcher instance from Config
func (watcher Watcher) Configure(config *Config, fwatcher *filewatcher.Watcher) error {
	//set the watcher to only watch
	fwatcher.FilterOps(filewatcher.Create, filewatcher.Remove, filewatcher.Write)

	if err := fwatcher.AddRecursive(config.NomadAllocDir); err != nil {
		return fmt.Errorf("Error watching Nomad alloc dir Error:%s", err.Error())
	}

	//Register watch Handlers
	watcher.registerWatchHandlers()

	return nil
}

func (watcher Watcher) registerWatchHandlers() {
	go func() {
		for {
			select {
			case event := <-watcher.filewatcher.Event:
				watcher.AllocationHandler.HandleAllocationEvents(event)
			case err := <-watcher.filewatcher.Error:
				watcher.ErrorEvents <- fmt.Errorf("Filewatcher Error:%s", err.Error())
			case <-watcher.filewatcher.Closed:
				return
			}
		}
	}()
}
