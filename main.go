package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/tebeka/selenium"
)

const (
	geckoDriverPath = "/usr/bin/geckodriver"
	port            = 4444
	wireLoginURL    = "https://app.wire.com/auth/#login"
)

func wireLogin(webDriver selenium.WebDriver, username, password string) (
	err error) {

	err = webDriver.Get(wireLoginURL)
	if err != nil {
		return
	}

	element, err := webDriver.FindElement("xpath", "//input[@name='email']")
	if err != nil {
		return
	}

	err = element.Clear()
	if err != nil {
		return
	}

	err = element.SendKeys(username)
	if err != nil {
		return
	}

	element, err = webDriver.FindElement("xpath", "//input[@name='password']")
	if err != nil {
		return
	}

	err = element.Clear()
	if err != nil {
		return
	}

	err = element.SendKeys(password)
	if err != nil {
		return
	}

	element, err = webDriver.FindElement("xpath", "//button[@type='submit']")
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	return nil
}

func wireRemoveClient(webDriver selenium.WebDriver, password string) (
	err error) {

	err = webDriver.WaitWithTimeout(func(wd selenium.WebDriver) (
		success bool, err error) {

		element, err := webDriver.FindElement("xpath",
			"//div[@data-uie-name='go-remove-device']")
		log.Printf("%v %v", err != nil, element)
		return err == nil && element != nil, nil
	}, 5*time.Second)

	if err != nil {
		return
	}

	element, err := webDriver.FindElement("xpath",
		"//div[@data-uie-name='go-remove-device']")
	if err != nil {
		return
	}

	element.Click()

	element, err = webDriver.FindElement("xpath", "//input[@name='password']")
	if err != nil {
		return
	}

	element.Click()
	element.SendKeys(password)
	element.SendKeys(selenium.EnterKey)

	return nil
}

func wireHistoryInfo(webDriver selenium.WebDriver) (err error) {
	err = webDriver.WaitWithTimeout(func(wd selenium.WebDriver) (
		success bool, err error) {

		element, err := webDriver.FindElement("xpath",
			"//button[@data-uie-name='do-history-confirm']")
		log.Printf("%v %v", err != nil, element)
		return err == nil && element != nil, nil
	}, 5*time.Second)

	if err != nil {
		return
	}

	element, err := webDriver.FindElement("xpath",
		"//button[@data-uie-name='do-history-confirm']")
	if err != nil {
		return
	}

	element.Click()

	return nil
}

func main() {
	caps := selenium.Capabilities{"browserName": "firefox"}
	var err error
	var webDriver selenium.WebDriver
	webDriver, err = selenium.NewRemote(caps,
		fmt.Sprintf("http://localhost:%d", port))

	if err != nil {
		log.Printf(err.Error())
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		log.Printf("Received interrupt")
		webDriver.Quit()
		log.Printf("Calling os.Exit(0)")
		os.Exit(0)
	}()

	defer webDriver.Quit()

	username := os.Getenv("WIRE_BOT_USERNAME")
	password := os.Getenv("WIRE_BOT_PASSWORD")

	err = wireLogin(webDriver, username, password)
	if err != nil {
		log.Printf("wireLogin error: %s", err.Error())
		return

	}

	for {
		var url string
		url, err = webDriver.CurrentURL()
		if err != nil {
			panic(err)
		}

		log.Printf("url = %s", url)

		switch url {
		case "https://app.wire.com/auth/#clients":
			wireRemoveClient(webDriver, password)
		case "https://app.wire.com/auth/#historyinfo":
			wireHistoryInfo(webDriver)
		case "https://app.wire.com/auth/#login":
			continue
		case "https://app.wire.com/":
			log.Printf("You made it!")
			time.Sleep(5 * time.Second)
			return
		default:
			log.Fatalf("Unrecognized url: %s", url)
		}
	}
}
