package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type serviceDetail struct {
	Logfile      string `json:"logfile"`
	Pattern      string `json:"pattern"`
	Action       action `json:"action"`
	TimeDuration int    `json:"timeDuration"`
	Enable       bool   `json:"enable"`
}

type action struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type service struct {
	Name          string        `json:"name"`
	ServiceDetail serviceDetail `json:"serviceDetail"`
}

func config() service {

	var configService service
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return configService
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configService)

	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return configService
	}

	fmt.Println(configService)
	return configService
}

func main() {

	configService := config()
	monitor(configService)
}

func monitor(service service) {

	filename := service.ServiceDetail.Logfile
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	var lastSize int64 = 0
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for {

		for scanner.Scan() {
			str := scanner.Text()
			re := regexp.MustCompile(service.ServiceDetail.Pattern)
			result := re.FindString(str)

			if result != "" {
				fmt.Println("Got an error, trying to heal -> ", service.Name)
				cmd := exec.Command(service.ServiceDetail.Action.Command, service.ServiceDetail.Action.Args...)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println("Error executing command:", err)
					return
				}
				fmt.Println(string(output))
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}

		// wait for a moment before checking for new changes
		time.Sleep(1 * time.Second)

		// check if the file has been updated
		fileinfo, err := os.Stat(filename)

		if err != nil {
			fmt.Println("Error getting file info:", err)
			return
		}

		size := fileinfo.Size()

		// check if the file size has increased, indicating new changes
		if size > lastSize {
			lastSize = size
			file.Seek(lastSize, int(size))
			scanner = bufio.NewScanner(file)
		}
	}
}
