package shepard

import (
	"log"

	"github.com/adragoset/shepard/watchers/alloc"

	consulapi "github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
)

//Config struct
type Config struct {
	Nomad        *nomadapi.Config     `json:"nomad"`
	Consul       *consulapi.Config    `json:"consul"`
	AllocWatcher *allocwatcher.Config `json:"alloc_watcher"`
}

//Shepard struct
type Shepard struct {
	consulClient *consulapi.Client
	nomadClient  *nomadapi.Client
	allocWatcher *allocwatcher.Watcher
}

//New Shepard
func New(config *Config) *Shepard {
	watcher, err := allocwatcher.New(config.AllocWatcher)
	if err != nil {
		log.Fatalf("Could Instantiate Allocation Watcher Error:%s", err.Error())
	}

	//cClient, err := consulapi.NewClient(config.Consul)
	//if err != nil {
	//log.Fatalf("Could not instantiate Concul Client Error:%s", err.Error())
	//}

	//nClient, err := nomadapi.NewClient(config.Nomad)
	//if err != nil {
	//log.Fatalf("Could not instantiate Nomad Client Error:%s", err.Error())
	//}

	//return &Shepard{consulClient: cClient, nomadClient: nClient, allocWatcher: watcher}

	return &Shepard{allocWatcher: watcher}
}

//Start Sheparding
func (sh Shepard) Start() {
	sh.allocWatcher.Start()
}

//Stop Sheparding
func (sh Shepard) Stop() {
	sh.allocWatcher.Stop()
}

func (sh Shepard) logErrors() {
	go func() {
		for {
			select {
			case err := <-sh.allocWatcher.ErrorEvents:
				log.Printf(err.Error())
			}
		}
	}()
}
