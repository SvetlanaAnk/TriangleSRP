package main

import (
	"fmt"
	"regexp"
)

func regexmatchzkill(link string) bool {
	match, err := regexp.MatchString("zkillboard.com/([0-9]+)", link)

	if err != nil {
		fmt.Println("Regex error", err)
		return false
	}
	return match
}
