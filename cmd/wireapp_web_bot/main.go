package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lk16/wireapp_web_bot"
	"github.com/tebeka/selenium"
	"log"
	"os"
	"os/signal"
)

const (
	port = 4444
)

func main() {

	var capabilitiesRaw string
	flag.StringVar(&capabilitiesRaw, "caps", "{}", "Browser capabilities")

	var seleniumHost string
	flag.StringVar(&seleniumHost, "host", "localhost", "Selenium host")

	var seleniumPort int
	flag.IntVar(&seleniumPort, "port", 4444, "Selenium port")

	var username string
	flag.StringVar(&username, "username", "", "Wireapp username")

	var password string
	flag.StringVar(&password, "password", "", "Wireapp password")

	flag.Parse()

	seleniumRemote := fmt.Sprintf("http://%s:%d", seleniumHost, seleniumPort)

	var capabilities map[string]interface{}

	err := json.Unmarshal([]byte(capabilitiesRaw), &capabilities)
	if err != nil {
		log.Fatalf(err.Error())
	}

	webDriver, err := selenium.NewRemote(capabilities, seleniumRemote)

	if err != nil {
		log.Fatalf(err.Error())
	}
	defer webDriver.Quit()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		log.Printf("Received interrupt")
		webDriver.Quit()
		log.Printf("Calling os.Exit(0)")
		os.Exit(0)
	}()

	_, err = wireapp.NewWireApp(webDriver, username, password)

	if err != nil {
		webDriver.Quit()
		log.Fatalf(err.Error())
	}

}
