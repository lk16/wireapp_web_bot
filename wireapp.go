package wireapp

import (
	"fmt"
	"log"
	"time"

	"github.com/tebeka/selenium"
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

	previousURL := ""

	for {
		time.Sleep(100 * time.Millisecond)

		var URL string
		URL, err = wa.webDriver.CurrentURL()
		if err != nil {
			return nil
		}

		if URL != previousURL {
			log.Printf("url = %s", URL)
		}

		previousURL = URL

		// TODO use error return values in switch

		switch URL {
		case "https://app.wire.com/auth/#clients":
			err = wa.pageAuthClients()
		case "https://app.wire.com/auth/#historyinfo":
			err = wa.pageAuthHistoryInfo()
		case "https://app.wire.com/auth/#login":
			continue
		case "https://app.wire.com/":
			return nil
		default:
			log.Fatalf("Unrecognized url: %s", URL)
		}
		if err != nil {
			switch err.(type) {
			case ChangedURLError:
				continue
			default:
				return
			}
		}
	}
}

func (wa *WireApp) pageAuthClients() (err error) {

	element, err := waitForElementXPath(wa.webDriver,
		"//div[@data-uie-name='go-remove-device']", 5*time.Second)

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

func (wa *WireApp) pageAuthHistoryInfo() (err error) {

	element, err := waitForElementXPath(wa.webDriver,
		"//button[@data-uie-name='do-history-confirm']", 5*time.Second)

	if err != nil {
		return
	}

	element.Click()

	return nil
}

type Conversation struct {
	element selenium.WebElement
	wireapp *WireApp
}

func (conv *Conversation) GetName() (name string, err error) {
	return conv.element.GetAttribute("data-uie-value")
}

func (conv *Conversation) SendMessage(message string) (err error) {
	element, err := conv.element.FindElement("xpath",
		"//div[@class='conversation-list-cell-center']")

	if err != nil {
		return
	}

	element.Click()

	textArea, err := conv.wireapp.webDriver.FindElement("xpath",
		"//textarea[@id='conversation-input-bar-text']")

	if err != nil {
		return
	}

	textArea.Click()
	textArea.SendKeys(message)
	textArea.SendKeys(selenium.EnterKey)

	return
}

func (wa *WireApp) ListConversations() (
	conversations []*Conversation, err error) {

	URL, err := wa.webDriver.CurrentURL()
	if err != nil {
		return
	}

	if URL != "https://app.wire.com/" {
		err = fmt.Errorf("Invalid URL: %s", URL)
		return
	}

	elements, err := wa.webDriver.FindElements("xpath",
		"//conversation-list-cell/div[contains(@class,"+
			"'conversation-list-cell')]")

	if err != nil {
		return
	}

	conversations = make([]*Conversation, len(elements))

	for i, element := range elements {
		conversations[i] = &Conversation{
			element: element,
			wireapp: wa}
	}

	return
}

func (wa *WireApp) FindConversation(targetName string) (
	conversation *Conversation, err error) {

	var conversations []*Conversation
	conversations, err = wa.ListConversations()
	if err != nil {
		return
	}

	for _, conversation := range conversations {

		var name string
		name, err = conversation.GetName()

		if err != nil {
			return nil, err
		}

		if name == targetName {
			return conversation, err
		}
	}

	err = fmt.Errorf("conversation not found")
	return
}
