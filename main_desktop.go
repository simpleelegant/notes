// +build !android

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/simpleelegant/notes/conf"
	"github.com/simpleelegant/notes/models"
)

func init() {
	host := flag.String("host", "127.0.0.1", "server host")
	port := flag.Int("port", 8080, "server port")

	// print usage
	fmt.Println("----------------------------------------")
	flag.Usage()
	fmt.Println("----------------------------------------")

	flag.Parse()

	conf.Host = *host
	conf.Port = *port
	conf.SetDataFolder(".")

	if info := conf.GatherInfo(); info.ErrorInMemory != "" {
		exit(errors.New(info.ErrorInMemory))
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	// init models
	if err := models.Init(conf.GetDataFilePath()); err != nil {
		exit(err)
	}

	registerRoutes(http.FileServer(http.Dir("./")))

	addr := conf.GatherInfo().ServerAddress
	fmt.Printf("Listening and serving HTTP on %s\n", addr)

	// Start the web server
	if err := http.ListenAndServe(addr, nil); err != nil {
		exit(err)
	}
}
