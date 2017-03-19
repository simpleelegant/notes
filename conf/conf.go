package conf

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

// configuration items
var (
	Host string
	Port int
)

// StartedAt server starting timestamp
var StartedAt = time.Now()

var (
	confFile     *os.File
	dataFilePath string
)

// GetHTTPAddress return the address at which HTTP serving
func GetHTTPAddress() string {
	return fmt.Sprintf("%s:%d", Host, Port)
}

// GetDataFolder get folder in which data file placed
func GetDataFolder() string {
	return dataFilePath
}

// SetDataFolder set where to store database file and configuration file
func SetDataFolder(df string) error {
	if !strings.HasSuffix(df, "/") {
		df += "/"
	}

	dataFilePath = df + "notes.db"

	// open configuration file
	f, err := os.OpenFile(df+"conf.dat", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	confFile = f

	return nil
}

// GetLastRestoringTimestamp return timestamp of last data restoring
func GetLastRestoringTimestamp() (string, error) {
	if _, err := confFile.Seek(0, 0); err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(confFile)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SetLastRestoringTimestamp set last data storing timestamp
func SetLastRestoringTimestamp() error {
	if _, err := confFile.Seek(0, 0); err != nil {
		return err
	}

	if _, err := confFile.WriteString(time.Now().String()); err != nil {
		return err
	}

	return nil
}

// GetComputerLocalIP return computer local ip
func GetComputerLocalIP() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}
