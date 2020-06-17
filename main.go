package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	APIKey     string = os.Getenv("APIKey")
	Zone       string = os.Getenv("Zone")
	RecordName string = os.Getenv("RecordName")

	tr = &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    5 * time.Second,
		DisableCompression: true,
	}

	previousIP string
)

func main() {
	log.Printf("Getting the ID of the record (%s)\n", Zone)
	recID, err := getRecordID()
	if err != nil {
		os.Exit(-1)
	}
	log.Printf("Got record ID: %s\n", recID)
	for {
		ip, err := getIP()
		if err == nil {
			if previousIP == ip {
				log.Println("No change to public IP")
			} else {
				log.Printf("Current public IP: %s\n", ip)
				err = setIP(ip, recID)
				if err != nil {
					log.Printf("Error setting IP: %v\n", err)
				} else {
					log.Println("Updated!")
					previousIP = ip
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

// DNSRecords holds the result of Cloudflare's DNS record listing.
type DNSRecords struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
		Meta struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
	} `json:"result"`
}

func setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+APIKey)
	req.Header.Set("Content-Type", "application/json")
}

func newClient() *http.Client {
	return &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}
}

// getRecordID gets the record ID of a DNS record based on the record name
func getRecordID() (string, error) {
	client := newClient()
	url := "https://api.cloudflare.com/client/v4/zones/" + Zone + "/dns_records?type=A&name=" + RecordName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %v", err)
	}
	setHeaders(req)
	resp, err := client.Do(req)
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudflare returned non-OK status: %s", resp.Status)
	}
	var recs DNSRecords
	err = json.Unmarshal(s, &recs)
	if len(recs.Result) == 0 {
		return "", fmt.Errorf("cannot find record: %s", RecordName)
	}

	return recs.Result[0].ID, nil
}

// setIP sets the IP address based on record ID.
func setIP(ip string, recID string) error {
	client := newClient()
	body := `{"type":"A","name":"` + RecordName + `","content":"` + ip + `","ttl":120,"proxied":false}`
	url := "https://api.cloudflare.com/client/v4/zones/" + Zone + "/dns_records/" + recID
	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create request: %v", err)
	}
	setHeaders(req)
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudflare returned non-OK status: %s", resp.Status)
	}
	return nil
}

// getIP returns the public IP address.
func getIP() (string, error) {
	client := newClient()
	req, err := http.NewRequest("GET", "https://icanhazip.com", nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(s)), nil
}
