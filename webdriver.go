package wireapp

import (
	"fmt"
	"github.com/tebeka/selenium"
	"time"
)

// ChangedURLError indicates the URL changed unexpectedly
type ChangedURLError string

func (err ChangedURLError) Error() string {
	return string(err)
}

func waitForElementXPath(webDriver selenium.WebDriver, xpath string,
	duration time.Duration) (element selenium.WebElement, err error) {

	expectedURL, err := webDriver.CurrentURL()
	if err != nil {
		return
	}

	condition := func(wd selenium.WebDriver) (success bool, err error) {

		URL, err := webDriver.CurrentURL()
		if err != nil {
			return
		}

		if URL != expectedURL {
			err = ChangedURLError(fmt.Sprintf("URL changed from '%s' to '%s'",
				expectedURL, URL))
			return
		}

		element, err = webDriver.FindElement("xpath", xpath)
		return err == nil && element != nil, nil
	}

	err = webDriver.WaitWithTimeout(condition, duration)
	return element, err
}
