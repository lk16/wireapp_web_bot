package wireapp

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
)

const (
	wireRootURL = "https://app.wire.com/"
)

// WireApp represents a wireapp session
type WireApp struct {
	webDriver selenium.WebDriver
	username  string
	password  string
}

// NewWireApp creats a new WireApp
func NewWireApp(webDriver selenium.WebDriver, username, password string) (
	wireapp *WireApp, err error) {

	if username == "" || password == "" {
		err = fmt.Errorf("username or password is not set properly")
		return
	}

	wireapp = &WireApp{
		webDriver: webDriver,
		username:  username,
		password:  password}

	if err = wireapp.login(); err != nil {
		err = errors.Wrap(err, "login error")
		return
	}

	if err = wireapp.pagesAfterLogin(); err != nil {
		err = errors.Wrap(err, "browsing error")
		return
	}

	return wireapp, nil
}

func (wa *WireApp) login() (err error) {

	err = wa.webDriver.Get(wireRootURL + "auth/#login")
	if err != nil {
		return errors.Wrap(err, "getting login url failed")
	}

	element, err := waitForElementXPath(wa.webDriver, "//input[@name='email']",
		5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "could not find email field")
	}

	err = element.Clear()
	if err != nil {
		return errors.Wrapf(err, "could not clear email field")
	}

	err = element.SendKeys(wa.username)
	if err != nil {
		return errors.Wrap(err, "could not send email keystrokes")
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//input[@name='password']")
	if err != nil {
		return errors.Wrap(err, "could not find password field")
	}

	err = element.Clear()
	if err != nil {
		return errors.Wrap(err, "could not clear password field")
	}

	err = element.SendKeys(wa.password)
	if err != nil {
		return errors.Wrap(err, "could not send password keystrokes")
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//button[@type='submit']")
	if err != nil {
		return errors.Wrap(err, "could not find submit button")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click on submit button")
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
			return errors.Wrap(err, "could not get current URL")
		}

		if URL != previousURL {
			log.Printf("url = %s", URL)
		}

		previousURL = URL

		// TODO use error return values in switch

		switch URL {
		case wireRootURL + "auth/#clients":
			err = wa.pageAuthClients()
			if err != nil {
				switch errors.Cause(err).(type) {
				case ChangedURLError:
					err = nil
					break
				default:
					log.Printf("error cause has type %T", errors.Cause(err))
				}
			}
		case wireRootURL + "auth/#historyinfo":
			err = wa.pageAuthHistoryInfo()
		case wireRootURL + "auth/#login":
			continue
		case wireRootURL:
			_, err = wa.webDriver.ExecuteScript(
				"document.getElementById('warnings').remove();",
				[]interface{}{})
			return errors.Wrap(err, "could not remove warnings div")
		default:
			errors.Wrapf(err, "unrecognized url %s", URL)
		}
		if err != nil {
			return errors.Wrapf(err, "browsing %s failed", URL)
		}
	}
}

func (wa *WireApp) pageAuthClients() (err error) {

	element, err := waitForElementXPath(wa.webDriver,
		"//div[@data-uie-name='go-remove-device']", 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not find remove device div")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click remove device div")
	}

	element, err = wa.webDriver.FindElement("xpath",
		"//input[@name='password']")
	if err != nil {
		return errors.Wrap(err, "could not find password field")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click password field")
	}

	err = element.SendKeys(wa.password)
	if err != nil {
		return errors.Wrap(err, "could not send password keystrokes")
	}

	err = element.SendKeys(selenium.EnterKey)
	if err != nil {
		return errors.Wrap(err, "could not send enter keystroke")
	}

	return nil
}

func (wa *WireApp) pageAuthHistoryInfo() (err error) {

	element, err := waitForElementXPath(wa.webDriver,
		"//button[@data-uie-name='do-history-confirm']", 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not find history confirm button")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click history confirm button")
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
		err = errors.Wrap(err, "could not get current URL")
		return
	}

	if URL != wireRootURL {
		err = errors.Wrapf(err, "Unexpected URL: '%s'", URL)
		return
	}

	xpath := fmt.Sprintf("//div[@data-uie-uid='%s']", conv.uuid)
	element, err := conv.wireapp.webDriver.FindElement("xpath", xpath)

	if err != nil {
		err = errors.Wrap(err, "could not find div for conversation")
		return
	}

	name, err = element.GetAttribute("data-uie-value")
	if err != nil {
		err = errors.Wrap(err,
			"could not get topic atrribute from conversation div")
		return
	}

	return name, nil
}

// SendMessage sends a message in a Conversation
func (conv *Conversation) SendMessage(message string) (err error) {

	xpath := fmt.Sprintf("//div[@data-uie-uid='%s']", conv.uuid)
	element, err := conv.wireapp.webDriver.FindElement("xpath", xpath)
	if err != nil {
		return errors.Wrap(err,
			"could not find topic attribute from conversation div")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click conversation div")
	}

	condition := func(wd selenium.WebDriver) (success bool, err error) {

		xpath := fmt.Sprintf("//div[@data-uie-uid='%s' and "+
			"contains(@class,'conversation-list-cell-active')]", conv.uuid)
		element, err = conv.wireapp.webDriver.FindElement("xpath", xpath)
		return err == nil && element != nil, nil
	}

	err = conv.wireapp.webDriver.WaitWithTimeout(condition, 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not switch conversation")
	}

	element, err = waitForElementXPath(conv.wireapp.webDriver,
		"//textarea[@id='conversation-input-bar-text']", 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not find message textarea")
	}

	err = element.Click()
	if err != nil {
		return errors.Wrap(err, "could not click message textarea")
	}

	err = element.SendKeys(message)
	if err != nil {
		return errors.Wrap(err, "could not send message keystrokes")
	}

	err = element.SendKeys(selenium.EnterKey)
	if err != nil {
		return errors.Wrap(err, "could not send enter keystroke")
	}

	return nil
}

// ListConversations lists each Conversation
func (wa *WireApp) ListConversations() (
	conversations []*Conversation, err error) {

	URL, err := wa.webDriver.CurrentURL()
	if err != nil {
		err = errors.Wrap(err, "could not get current URL")
		return
	}

	if URL != wireRootURL {
		err = fmt.Errorf("Invalid URL: %s", URL)
		return
	}

	xpath := "//conversation-list-cell/div[contains(@class," +
		"'conversation-list-cell')]"

	_, err = waitForElementXPath(wa.webDriver, xpath, 5*time.Second)
	if err != nil {
		err = errors.Wrap(err, "could not find any conversation")
		return
	}

	elements, err := wa.webDriver.FindElements("xpath", xpath)
	if err != nil {
		err = errors.Wrap(err, "could not list any conversations")
		return
	}

	conversations = make([]*Conversation, len(elements))

	for i, element := range elements {
		conversations[i] = &Conversation{wireapp: wa}
		conversations[i].uuid, err = element.GetAttribute("data-uie-uid")
		if err != nil {
			err = errors.Wrap(err,
				"could not get uuid attribute of conversation")
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
		err = errors.Wrap(err, "could not list all conversations")
		return
	}

	for _, conversation := range conversations {

		var topic string
		topic, err = conversation.GetTopic()

		if err != nil {
			err = errors.Wrap(err, "could not get topic of conversation")
			return nil, err
		}

		if topic == targetTopic {
			return conversation, err
		}
	}

	err = errors.Wrap(err, "could not find conversation")
	return nil, err
}
