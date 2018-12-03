package common

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func DetermineHostID(hosts []string) int {
	myHostname, err := os.Hostname()
	if err != nil {
		log.Error("Could determine hostname: " + err.Error())
		os.Exit(1)
	}

	log.Info("Local hostame=" + myHostname)

	// determine pid from hostsfile
	var myHostID int
	for i, h := range hosts {
		if h == myHostname {
			myHostID = i
		}
	}
	log.Info("Local hostame=" + myHostname + "; host ID=" + strconv.Itoa(myHostID))
	return myHostID
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func ReadHostsFile(filename string) []string {
	hosts, err := readLines(filename)
	if err != nil {
		log.Error("Could not read hostsfile: " + err.Error())
		os.Exit(1)
	}
	log.Info("Hosts in hostsfile=" + strings.Join(hosts, " "))

	return hosts
}
