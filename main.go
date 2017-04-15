package main

import (
	"fmt"
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/halo779/broadcombeat/beater"
)

func main() {
	fmt.Println("Starting Broadcom Beat")
	err := beat.Run("broadcombeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}

}
