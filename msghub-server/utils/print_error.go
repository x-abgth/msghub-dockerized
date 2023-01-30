package utils

import "log"

func PrintError(err error, info string) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
