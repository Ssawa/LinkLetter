package main

import (
	"fmt"

	"github.com/Ssawa/LinkLetter/config"
)

func main() {
	test := config.ParseForConfig()
	fmt.Println(test)
}
