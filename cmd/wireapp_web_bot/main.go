package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lk16/wireapp_web_bot"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	var capabilitiesRaw string
	flag.StringVar(&capabilitiesRaw, "caps", "{}", "Browser capabilities")

	var seleniumHost string
	flag.StringVar(&seleniumHost, "host", "localhost", "Selenium host")

	var seleniumPort int
	flag.IntVar(&seleniumPort, "port", 4444, "Selenium port")

	var username string
	flag.StringVar(&username, "user", "", "Wireapp username")

	var password string
	flag.StringVar(&password, "pass", "", "Wireapp password")

	var topic string
	flag.StringVar(&topic, "topic", "", "Wireapp conversation topic")

	var headless bool
	flag.BoolVar(&headless, "headless", false, "Run headless mode")

	flag.Parse()

	seleniumRemote := fmt.Sprintf("http://%s:%d", seleniumHost, seleniumPort)

	capabilities := selenium.Capabilities{}

	if headless {
		capabilities.AddFirefox(firefox.Capabilities{
			Args: []string{"-headless"}})
	}

	err := json.Unmarshal([]byte(capabilitiesRaw), &capabilities)
	if err != nil {
		log.Fatalf(err.Error())
	}

	webDriver, err := selenium.NewRemote(capabilities, seleniumRemote)

	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func() {
		err = webDriver.Quit()
		if err != nil {
			return
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		log.Printf("Received interrupt")
		err = webDriver.Quit()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("Calling os.Exit(0)")
		os.Exit(0)
	}()

	wa, err := wireapp.NewWireApp(webDriver, username, password)

	if err != nil {
		log.Fatalf(err.Error())
	}

	conversation, err := wa.FindConversation(topic)
	if err != nil {
		log.Fatalf(err.Error())
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		err = conversation.SendMessage(message)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	time.Sleep(2 * time.Second)
	log.Printf("main() returns")
}
