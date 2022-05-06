package main

import "fmt"

func main() {

	k := K8SHandler{}
	k.init()
	k.GetPods()
	//k.GetDeployments()

	//k.SetCommand("travis-api", []string{"/bin/sh", "-c", "tail -f /dev/null"}, []string{})
}
