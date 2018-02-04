package main

import (
	"NomadShepard"
)

func main() {

	shepardConfig := shepard.LoadShepardConfiguration()

	shep := shepard.New(shepardConfig)

	shep.Start()

}
