package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Host struct {
	ID        string   `json:"id"`
	IPAddress string   `json:"ipAddress"`
	Hostname  string   `json:"name"`
	Tags      []string `json:"tags"`
}

type hostsResponse struct {
	Data     []Host `json:"data"`
	Metadata struct {
		HasNextPage bool   `json:"hasNextPage"`
		Cursor      string `json:"cursor"`
	} `json:"metadata"`
}

func FilterHosts(dnToken string, filterFunc func(Host) bool) ([]Host, error) {
	hosts := []Host{}

	cursor := ""
	for {
		// Fetch the next page of hosts
		params := url.Values{
			"cursor":   []string{cursor},
			"pageSize": []string{"500"},
		}

		req, err := http.NewRequest("GET", "https://api.defined.net/v1/hosts?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+dnToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
		}

		// Decode the response body into a slice of Hosts
		var respHosts hostsResponse
		err = json.Unmarshal(body, &respHosts)
		if err != nil {
			return nil, err
		}

		// Filter the hosts
		for _, host := range respHosts.Data {
			if filterFunc(host) {
				hosts = append(hosts, host)
			}
		}

		// Fetch the next page if there is one
		if !respHosts.Metadata.HasNextPage {
			break
		}
		cursor = respHosts.Metadata.Cursor
	}

	return hosts, nil
}
