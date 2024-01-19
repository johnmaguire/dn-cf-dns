package main

import "fmt"

type Record struct {
	ID   string
	Name string
}

func IterateRecords(cfToken string, zoneID string, fn func(record Record) error) error {
	return fmt.Errorf("not implemented")
}

func CreateRecord(cfToken string, zoneID string, hostname string, ip string) error {
	return fmt.Errorf("not implemented")
}

func DeleteRecord(cfToken string, zoneID string, recordID string) error {
	return fmt.Errorf("not implemented")
}
