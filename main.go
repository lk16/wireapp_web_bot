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
)

type WireApp struct {
	webDriver selenium.WebDriver
	username  string
	password  string
}

func NewWireApp(webDriver selenium.WebDriver, username, password string) (
	wa *WireApp, err error) {

	if username == "" || password == "" {
		err = fmt.Errorf("username or password is not set properly")
		return
	}

	wa = &WireApp{
		webDriver: webDriver,
		username:  username,
		password:  password}

	if err = wa.login(); err != nil {
		err = fmt.Errorf("login error: %s", err.Error())
		return
	}

	if err = wa.pagesAfterLogin(); err != nil {
		err = fmt.Errorf("pages after login error: %s", err.Error())
		return
	}

	return wa, nil
}

func (wa *WireApp) login() (err error) {

	err = wa.webDriver.Get("https://app.wire.com/auth/#login")
	if err != nil {
		return
	}

	element, err := wa.webDriver.FindElement("xpath", "//input[@name='email']")
	if err != nil {
		return
	}

	err = element.Clear()
	if err != nil {
		return
	}

	err = element.SendKeys(wa.username)
	if err != nil {
		return
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//input[@name='password']")
	if err != nil {
		return
	}

	err = element.Clear()
	if err != nil {
		return
	}

	err = element.SendKeys(wa.password)
	if err != nil {
		return
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//button[@type='submit']")
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	return nil
}

func (wa *WireApp) pagesAfterLogin() (err error) {
	for {
		time.Sleep(100 * time.Millisecond)

		url, err := wa.webDriver.CurrentURL()
		if err != nil {
			return nil
		}

		log.Printf("url = %s", url)

		switch url {
		case "https://app.wire.com/auth/#clients":
			wa.pageAuthClients()
		case "https://app.wire.com/auth/#historyinfo":
			wa.pageAuthhistoryInfo()
		case "https://app.wire.com/auth/#login":
			continue
		case "https://app.wire.com/":
			return nil
		default:
			log.Fatalf("Unrecognized url: %s", url)
		}
	}
}

func (wa *WireApp) pageAuthClients() (err error) {

	err = wa.webDriver.WaitWithTimeout(func(wd selenium.WebDriver) (
		success bool, err error) {

		element, err := wa.webDriver.FindElement("xpath",
			"//div[@data-uie-name='go-remove-device']")
		log.Printf("%v %v", err != nil, element)
		return err == nil && element != nil, nil
	}, 5*time.Second)

	if err != nil {
		return
	}

	element, err := wa.webDriver.FindElement("xpath",
		"//div[@data-uie-name='go-remove-device']")
	if err != nil {
		return
	}

	element.Click()

	element, err = wa.webDriver.FindElement("xpath",
		"//input[@name='password']")

	if err != nil {
		return
	}

	element.Click()
	element.SendKeys(wa.password)
	element.SendKeys(selenium.EnterKey)

	return nil
}

func (wa *WireApp) pageAuthhistoryInfo() (err error) {
	err = wa.webDriver.WaitWithTimeout(func(wd selenium.WebDriver) (
		success bool, err error) {

		element, err := wa.webDriver.FindElement("xpath",
			"//button[@data-uie-name='do-history-confirm']")
		log.Printf("%v %v", err != nil, element)
		return err == nil && element != nil, nil
	}, 5*time.Second)

	if err != nil {
		return
	}

	element, err := wa.webDriver.FindElement("xpath",
		"//button[@data-uie-name='do-history-confirm']")
	if err != nil {
		return
	}

	element.Click()

	return nil
}

func main() {
	caps := selenium.Capabilities{"browserName": "firefox"}
	webDriver, err := selenium.NewRemote(caps,
		fmt.Sprintf("http://localhost:%d", port))

	if err != nil {
		log.Printf(err.Error())
		return
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

	username := os.Getenv("WIRE_BOT_USERNAME")
	password := os.Getenv("WIRE_BOT_PASSWORD")

	_, err = NewWireApp(webDriver, username, password)

	if err != nil {
		panic(err)
	}

}
