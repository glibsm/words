package main

import (
	"log"

	"github.com/glibsm/words"
)

func main() {
	err := words.Serve(
		"Bloggy McBlogface",
		words.Port(8080),
	)
	if err != nil {
		log.Fatal(err)
	}
}
