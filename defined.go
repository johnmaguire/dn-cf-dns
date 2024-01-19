package main

import "fmt"

type Host struct {
	ID        string
	IPAddress string
	Hostname  string
	Tags      []string
}

func FilterHosts(dnToken string, filterFunc func(Host) bool) ([]Host, error) {
	return nil, fmt.Errorf("not implemented")
}
