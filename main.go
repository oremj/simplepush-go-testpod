package main

import "github.com/oremj/simplepush-go-testpod/testpod"

func main() {
	plan := testpod.NewPlan()
	plan.NumChannels = 2
	plan.NumClients = 2
	plan.URLs = []string{"wss://push.stage.mozaws.net"}
	plan.Go()
}
