package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

type Record struct {
	ID   string
	Name string
}

func GetZoneID(cf *cloudflare.API, zoneName string) (string, error) {
	zones, err := cf.ListZones(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	for _, z := range zones {
		if z.Name == zoneName {
			return z.ID, nil
		}
	}

	return "", fmt.Errorf("zone %s not found", zoneName)
}

func IterateRecords(cf *cloudflare.API, zoneID string, fn func(record Record) error) error {
	recs, _, err := cf.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %w", err)
	}

	for _, r := range recs {
		r := Record{ID: r.ID, Name: r.Name}
		if err := fn(r); err != nil {
			// TODO better error handling
			return fmt.Errorf("error in callback for record %+v: %w", r, err)
		}
	}

	return nil
}

func CreateRecord(cf *cloudflare.API, zoneID string, hostname string, ip string) error {
	// Check if the record already exists
	recs, _, err := cf.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Name: hostname,
	})
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %w", err)
	}

	if len(recs) > 0 {
		// update the record
		_, err := cf.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
			ID:      recs[0].ID,
			Type:    "A",
			Name:    hostname,
			Content: ip,
			TTL:     1,
			Proxied: cloudflare.BoolPtr(false),
		})
		if err != nil {
			return fmt.Errorf("failed to update DNS record: %w", err)
		}
	} else {
		_, err := cf.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
			Type:    "A",
			Name:    hostname,
			Content: ip,
			TTL:     1,
			Proxied: cloudflare.BoolPtr(false),
		})
		if err != nil {
			return fmt.Errorf("failed to create DNS record: %w", err)
		}
	}

	return nil
}

func DeleteRecord(cf *cloudflare.API, zoneID string, recordID string) error {
	err := cf.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}
