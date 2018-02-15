package allocwatcher

import (
	"fmt"
	"strings"
	"time"

	"github.com/adragoset/shepard/watchers/alloc/handlers"
	"github.com/adragoset/shepard/watchers/file"
)

//Config struct
type Config struct {
	NomadAllocDir string        `json:"nomad_alloc_dir"`
	WatchCycleMs  time.Duration `json:"watch_cycle_ms"`
}

//Watcher struct
type Watcher struct {
	Handlers          map[string]*allochandlers.FileWatchHandler
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

	allocWatcher = &Watcher{AllocationEvents: allocEventChannel, ErrorEvents: errorEventChannel, WatchCycleMs: config.WatchCycleMs, filewatcher: fileWatcher}

	// Watch the alloc folder for changes.
	err := allocWatcher.Configure(config)
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
func (watcher *Watcher) Configure(config *Config) error {
	//set the watcher to only watch
	watcher.filewatcher.FilterOps(filewatcher.Create, filewatcher.Remove, filewatcher.Write)

	if err := watcher.filewatcher.Add(config.NomadAllocDir); err != nil {
		return fmt.Errorf("Error watching Nomad alloc dir Error:%s", err.Error())
	}

	if err := watcher.buildHandlers(config.NomadAllocDir); err != nil {
		return fmt.Errorf("Error building watch handlers Error:%s", err.Error())
	}

	//Register watch Handlers
	watcher.registerWatchHandlers()

	return nil
}

func (watcher *Watcher) buildHandlers(allocDir string) error {
	allocHandler, err := allochandlers.NewAllocationHandler(allocDir, watcher.AllocationEvents, watcher.ErrorEvents)
	if err != nil {
		err = fmt.Errorf("AllocationWatcher failed to create an AllocationHandler Error:%s", err.Error())
		return err
	}

	watcher.filewatcher.IgnoreRegexs(strings.Join(allocHandler.FilterExpressions, ","))

	watcher.AllocationHandler = allocHandler
	return nil
}

func (watcher *Watcher) registerWatchHandlers() {
	go func() {
		for {
			select {
			case event := <-watcher.filewatcher.Event:
				watcher.AllocationHandler.HandleEvents(event)
			case err := <-watcher.filewatcher.Error:
				watcher.ErrorEvents <- fmt.Errorf("Filewatcher Error:%s", err.Error())
			case <-watcher.filewatcher.Closed:
				return
			}
		}
	}()
}
