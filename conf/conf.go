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

// Info system information
type Info struct {
	StartAt                   time.Time
	ServerAddress             string
	ComputerLocalIP           string
	RecentRestoredDataAt      string
	RecentRestoredDataVersion string

	// store error info if no suitable place to store it
	ErrorInMemory string
}

var (
	confFile     *os.File
	dataFilePath string
	info         Info
)

// SetDataFolder set where to store database file and configuration file
func SetDataFolder(df string) {
	if !strings.HasSuffix(df, "/") {
		df += "/"
	}

	dataFilePath = df + "notes.db"

	// open configuration file
	f, err := os.OpenFile(df+"conf.dat", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		info.ErrorInMemory = err.Error()
		return
	}
	confFile = f

	// read configuration file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		info.RecentRestoredDataAt = err.Error()
	}
	info.RecentRestoredDataAt = string(b)
}

// GetDataFilePath return database file path
func GetDataFilePath() string {
	return dataFilePath
}

// GatherInfo gather system information
func GatherInfo() Info {
	if info.StartAt.IsZero() {
		info.StartAt = time.Now()
	}

	if info.ServerAddress == "" {
		info.ServerAddress = fmt.Sprintf("%s:%d", Host, Port)
	}

	// get local ip
	ips, err := getLocalIP()
	if err != nil {
		info.ComputerLocalIP = err.Error()
	} else {
		info.ComputerLocalIP = strings.Join(ips, ", ")
	}

	return info
}

// FreshRecentRestoredDataAt set RecentRestoredDataAt to current time
func FreshRecentRestoredDataAt() {
	info.RecentRestoredDataAt = time.Now().String()

	if confFile != nil {
		if _, err := confFile.Seek(0, 0); err != nil {
			info.ErrorInMemory = err.Error()
			return
		}

		if _, err := confFile.WriteString(info.RecentRestoredDataAt); err != nil {
			info.ErrorInMemory = err.Error()
		}
	}
}

// get local IPs
func getLocalIP() (ips []string, e error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return
}
