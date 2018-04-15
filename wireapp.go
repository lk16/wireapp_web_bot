package wireapp

import (
	"fmt"
	"log"
	"time"

	"github.com/tebeka/selenium"
)

// WireApp represents a wireapp session
type WireApp struct {
	webDriver selenium.WebDriver
	username  string
	password  string
}

// NewWireApp creats a new WireApp
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

	element, err := waitForElementXPath(wa.webDriver, "//input[@name='email']",
		5*time.Second)
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

	err = element.Click()
	if err != nil {
		return
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//input[@name='password']")
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	err = element.SendKeys(wa.password)
	if err != nil {
		return
	}

	err = element.SendKeys(selenium.EnterKey)
	if err != nil {
		return
	}

	return nil
}

func (wa *WireApp) pageAuthHistoryInfo() (err error) {

	element, err := waitForElementXPath(wa.webDriver,
		"//button[@data-uie-name='do-history-confirm']", 5*time.Second)
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	return nil
}

// Conversation represents a wireapp conversation
type Conversation struct {
	uuid    string
	wireapp *WireApp
}

// GetTopic gets the topic of a Conversation
func (conv *Conversation) GetTopic() (name string, err error) {

	URL, err := conv.wireapp.webDriver.CurrentURL()
	if err != nil {
		return
	}

	if URL != "https://app.wire.com/" {
		err = fmt.Errorf("Unexpected URL: '%s'", URL)
		return
	}

	xpath := fmt.Sprintf("//div[@data-uie-uid='%s']", conv.uuid)
	element, err := conv.wireapp.webDriver.FindElement("xpath", xpath)

	if err != nil {
		return
	}

	return element.GetAttribute("data-uie-value")
}

// SendMessage sends a message in a Conversation
func (conv *Conversation) SendMessage(message string) (err error) {

	xpath := fmt.Sprintf("//div[@data-uie-uid='%s']", conv.uuid)
	element, err := conv.wireapp.webDriver.FindElement("xpath", xpath)
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	condition := func(wd selenium.WebDriver) (success bool, err error) {

		xpath := fmt.Sprintf("//div[@data-uie-uid='%s' and "+
			"contains(@class,'conversation-list-cell-active')]", conv.uuid)
		element, err = conv.wireapp.webDriver.FindElement("xpath", xpath)
		if err != nil {
			log.Printf(err.Error())
		}
		return err == nil && element != nil, nil
	}

	err = conv.wireapp.webDriver.WaitWithTimeout(condition, 5*time.Second)
	if err != nil {
		return
	}

	element, err = waitForElementXPath(conv.wireapp.webDriver,
		"//textarea[@id='conversation-input-bar-text']", 5*time.Second)
	if err != nil {
		return
	}

	err = element.Click()
	if err != nil {
		return
	}

	err = element.SendKeys(message)
	if err != nil {
		return
	}

	err = element.SendKeys(selenium.EnterKey)
	if err != nil {
		return
	}

	return err
}

// ListConversations lists each Conversation
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

	xpath := "//conversation-list-cell/div[contains(@class," +
		"'conversation-list-cell')]"

	_, err = waitForElementXPath(wa.webDriver, xpath, 5*time.Second)
	if err != nil {
		return
	}

	elements, err := wa.webDriver.FindElements("xpath", xpath)

	if err != nil {
		return
	}

	conversations = make([]*Conversation, len(elements))

	for i, element := range elements {
		conversations[i] = &Conversation{wireapp: wa}
		conversations[i].uuid, err = element.GetAttribute("data-uie-uid")
		if err != nil {
			return
		}
	}

	return conversations, nil
}

// FindConversation finds a Conversation by topic
func (wa *WireApp) FindConversation(targetTopic string) (
	conversation *Conversation, err error) {

	var conversations []*Conversation
	conversations, err = wa.ListConversations()
	if err != nil {
		return
	}

	for _, conversation := range conversations {

		var topic string
		topic, err = conversation.GetTopic()

		if err != nil {
			return nil, err
		}

		if topic == targetTopic {
			return conversation, err
		}
	}

	err = fmt.Errorf("conversation not found")
	return nil, err
}
