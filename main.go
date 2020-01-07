package main

import (
	"cadence_helloworld/cadence"
	"flag"
)



func flow(process string) {
	switch process {
	case "wf":
		cadence.StartCadenceWorker()
	default:
		cadence.StartWorkflow()
	}
}

func main() {
	var process string
	flag.StringVar(&process, "process", "default", "Start the process")
	flag.Parse()
	flow(process)
}
