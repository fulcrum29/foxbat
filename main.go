package main

import (
	"log"

	promethues "github.com/prometheus/client_golang/api"
)

func main() {

	promConf := promethues.Config{
		Address: "localhost:9000",
	}

	promClient, err := promethues.NewClient(promConf)

	if err != nil {
		log.Fatal(err)
	}

}
