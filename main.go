package main

import (
	"os"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/stanchan/mattermostbeat/beater"
)

func main() {
	err := beat.Run("mattermostbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
