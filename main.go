package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
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
		logrus.Error("Couldn't get DNS records")
		logrus.Error(err)
		return err
	}

	for _, record := range records {
		if record.Name == targetRecord {
			record.Content = ipAddress
			err := api.UpdateDNSRecord(id, record.ID, record)
			if err != nil {
				return errors.New("Failed to update DNS record")
			}
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Couldn't find target record. Choices are: %s", records))
}

func getIPv4() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
