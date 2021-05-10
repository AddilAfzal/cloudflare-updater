package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	ApiKey           = flag.String("apiKey", "", "is the CloudFlare API key")
	Email            = flag.String("email", "", "is the account email")
	TargetZone       = flag.String("targetZone", "", "is zone name in which record is found")
	TargetRecordName = flag.String("targetRecord", "", "is record name to update")
)

func init() {
	flag.Parse()
}

func main() {
	var (
		cacheIPv4 string
		s         = make(chan os.Signal)
	)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	logrus.Info("Starting")
	logrus.Infof("Target zone: %s", *TargetZone)
	logrus.Infof("Target host record: %s", *TargetRecordName)
	logrus.Infof("Email address: %s", *Email)
	logrus.Infof("API key: %s", *ApiKey)

	c := cron.New()

	c.AddFunc("0 */1 * * * *", func() {
		currentIPv4, err := getIPv4()
		if err != nil {
			logrus.Errorf("Couldn't get current IP")
			logrus.Error(err)
			return
		}

		if currentIPv4 != cacheIPv4 {
			logrus.Infof("New IP: %s", currentIPv4)
			logrus.Info("Updating Cloudflare")
			err := UpdateCloudflareRecord(*TargetRecordName, currentIPv4)
			cacheIPv4 = currentIPv4
			if err != nil {
				logrus.Error(err)
			}
		}
		return
	})

	c.Start()
	<-s
	logrus.Info("Stopping")
	c.Stop()
}

func UpdateCloudflareRecord(targetRecord, ipAddress string) error {

	// Construct a new API object
	api, err := cloudflare.New(*ApiKey, *Email)
	if err != nil {
		return err
	}

	// Fetch the zone ID
	id, err := api.ZoneIDByName(*TargetZone)
	if err != nil {
		return err
	}

	records, err := api.DNSRecords(id, cloudflare.DNSRecord{})

	if err != nil {
		return err
	}

	var recordValues []string

	for _, record := range records {
		recordValues = append(recordValues, record.Name)
		if record.Name == targetRecord {
			record.Content = ipAddress
			err := api.UpdateDNSRecord(id, record.ID, record)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Couldn't find target record. Options are: %s", strings.Join(recordValues, ", ")))
}

func getIPv4() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if net.ParseIP(string(body)) == nil {
		return "", errors.New(fmt.Sprintf("Invalid IP Address: %s\n", string(body)))
	}

	return string(body), nil
}
