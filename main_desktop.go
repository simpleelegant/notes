// +build !android

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/simpleelegant/notes/conf"
	"github.com/simpleelegant/notes/resources"
)

func init() {
	host := flag.String("host", "127.0.0.1", "server host")
	port := flag.Int("port", 9030, "server port")

	// print usage
	fmt.Println("----------------------------------------")
	flag.Usage()
	fmt.Println("----------------------------------------")

	flag.Parse()

	conf.Host = *host
	conf.Port = *port

	if err := conf.SetDataFolder("."); err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	if err := resources.OpenDatabase(conf.GetDataFolder()); err != nil {
		exit(err)
	}

	registerRoutes(http.FileServer(http.Dir("./")))

	addr := conf.GetHTTPAddress()
	fmt.Printf("Listening and serving HTTP on %s\n", addr)

	// Start the web server
	if err := http.ListenAndServe(addr, nil); err != nil {
		exit(err)
	}
}
