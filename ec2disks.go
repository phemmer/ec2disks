package main

import (
	"flag"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Disk struct {
	Name    string
	Type    string
	Id      string
	Aliases []string
}

var version string
var buildDate string

func httpGet(path string) (content string, err error) {
	url := "http://169.254.169.254/" + path
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	content = string(bytes)
	return content, nil
}

func main() {
	flagDeviceNoPrefix := flag.Bool("s", false, "Output device name instead of path")
	flagAliases := flag.Bool("A", false, "Output aliases only")
	flagVersion := flag.Bool("V", false, "Show version information")
	flag.Parse()

	if *flagVersion {
		fmt.Printf("ec2disks %s, built %s\n", version, buildDate)
		os.Exit(0)
	}

	devicePrefix := ""
	if !*flagDeviceNoPrefix {
		devicePrefix = "disk/ec2/"
	}

	search := ""
	if len(flag.Args()) == 1 {
		search = flag.Arg(0)
		if !strings.HasPrefix(search, "/dev/") {
			search = "/dev/" + search
		}
	}

	disks := make(map[string]*Disk)

	// get the instance ID
	instanceId := aws.InstanceId()
	if instanceId == "unknown" {
		fmt.Fprintf(os.Stderr, "Could not get instance ID\n")
		os.Exit(1)
	}

	// first look in the metadata service
	devicesString, err := httpGet("2014-02-25/meta-data/block-device-mapping")
	if err == nil {
		deviceMapNames := strings.Split(devicesString, "\n")
		for _, deviceMapName := range deviceMapNames {
			deviceName, err := httpGet("2014-02-25/meta-data/block-device-mapping/" + deviceMapName)
			if !strings.HasPrefix(deviceName, "/dev/") {
				deviceName = "/dev/" + deviceName
			}
			deviceName = strings.Replace(deviceName, "/dev/sd", "/dev/xvd", 1)
			if err != nil {
				continue
			}
			disk := disks[deviceName]
			if disk == nil {
				disk = &Disk{Name: deviceName}
				disks[deviceName] = disk
			}
			disk.Aliases = append(disk.Aliases, devicePrefix+deviceMapName)
			if strings.HasPrefix(deviceMapName, "ephemeral") || deviceMapName == "swap" {
				disk.Type = "ephemeral"
			}
		}
	}

	// then look up using an API call
	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get AWS credentials\n")
		os.Exit(1)
	}
	e := ec2.New(auth, aws.Regions[aws.InstanceRegion()])

	resp, err := e.DescribeInstances([]string{instanceId}, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get block device mapping from AWS (%s)\n", err.Error())
		os.Exit(1)
	}

	instance := resp.Reservations[0].Instances[0]

	for _, device := range instance.BlockDevices {
		name := device.DeviceName
		if ! strings.HasPrefix(name, "/") {
			name = "/dev/" + name
		}
		name = strings.Replace(name, "/dev/sd", "/dev/xvd", 1)

		disk := disks[name]
		if disk == nil {
			disk = &Disk{Name: name}
			disks[name] = disk
		}
		disk.Type = "ebs"
		disk.Id = device.EBS.VolumeId
		disk.Aliases = append(disk.Aliases, devicePrefix+device.EBS.VolumeId)
	}

	for _, disk := range disks {
		if search != "" && disk.Name != search {
			continue
		}

		if *flagAliases {
			if search == "" {
				if !*flagDeviceNoPrefix {
					fmt.Printf("%s ", disk.Name)
				} else {
					fmt.Printf("%s ", strings.Replace(disk.Name, "/dev/", "", 1))
				}
			}
			fmt.Println(strings.Join(disk.Aliases, " "))
		} else {
			fmt.Printf("%s: TYPE=%s ID=%s ALIASES=%s\n", disk.Name, disk.Type, disk.Id, strings.Join(disk.Aliases, ","))
		}

		if search != "" {
			os.Exit(0)
		}
	}

	if search != "" {
		os.Exit(1)
	}

	//fmt.Println("%#v", instance.BlockDevices)
}
