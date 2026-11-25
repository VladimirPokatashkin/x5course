package config

import "os"

func GetAddress() string {
	var address = os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	return address
}
