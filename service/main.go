package main

import (
	"github.com/adragoset/shepard"
)

func main() {

	shepardConfig := shepard.LoadShepardConfiguration()

	shep := shepard.New(shepardConfig)

	shep.Start()

}
