/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package main

import (
	"os"

	"github.com/score-spec/score-humanitec/internal/command"
)

func main() {

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
