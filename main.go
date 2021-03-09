package main

import (
	"fmt"

	"github.com/ragadeeshu/octo-canyon-freshness-league/datahandling"
)

func main() {
	league, err := datahandling.LoadLeague()
	if err != nil {
		fmt.Println(err)
	}
	err = datahandling.LoadSplatnetData(&league)
	if err != nil {
		fmt.Println(err)
	}
	err = datahandling.SaveLeague(league)
	if err != nil {
		fmt.Println(err)
	}

}
