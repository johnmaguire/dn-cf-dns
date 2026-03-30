package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
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

func dnRequest(dnToken string, method string, path string, query url.Values) (*http.Response, error) {
	url := fmt.Sprintf("https://api.defined.net%s?%s", path, query.Encode())
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+dnToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code: %d, failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	return resp, nil
}

func GetNetworkCIDR(dnToken string, networkID string) (netip.Prefix, error) {
	resp, err := dnRequest(dnToken, "GET", fmt.Sprintf("/v1/networks/%s", networkID), nil)
	if err != nil {
		return netip.Prefix{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return netip.Prefix{}, err
	}

	var res struct {
		Data struct {
			CIDR string `json:"cidr"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return netip.Prefix{}, err
	}

	return netip.ParsePrefix(res.Data.CIDR)
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

		resp, err := dnRequest(dnToken, "GET", "/v1/hosts", params)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
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
