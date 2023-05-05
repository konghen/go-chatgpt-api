package webdriver

import (
	"strings"
	"time"

	"github.com/linweiyuan/go-chatgpt-api/api"
	"github.com/tebeka/selenium"
)

const (
	checkCaptchaTimeout      = 15
	checkCaptchaInterval     = 1
	clickedCaptchaWaitSecond = 5
)

func isReady(webDriver selenium.WebDriver) bool {
	err := webDriver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		title, _ := driver.Title()
		if strings.Contains(title, api.ChatGPTTitleText) {
			return true, nil
		}

		return false, nil
	}, time.Second*checkCaptchaTimeout, time.Second*checkCaptchaInterval)

	return err == nil
}

//goland:noinspection GoUnhandledErrorResult
func HandleCaptcha(webDriver selenium.WebDriver) bool {
	webDriver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		title, _ := driver.Title()
		if strings.Contains(title, api.ChatGPTTitleText) {
			return true, nil
		}

		if err := webDriver.SwitchFrame(0); err != nil {
			return false, nil
		}

		return true, nil
	}, time.Second*checkCaptchaTimeout, time.Second*checkCaptchaInterval)

	title, _ := webDriver.Title()
	if strings.Contains(title, api.ChatGPTTitleText) {
		return true
	}

	err := webDriver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		title, _ := webDriver.Title()
		if strings.Contains(title, api.ChatGPTTitleText) {
			return true, nil
		}

		element, err := driver.FindElement(selenium.ByCSSSelector, "input")
		if err != nil {
			return false, nil
		}

		element.Click()
		time.Sleep(time.Second * clickedCaptchaWaitSecond)
		return true, nil
	}, time.Second*checkCaptchaTimeout, time.Second*checkCaptchaInterval)

	if err != nil {
		retry(webDriver)
	} else {
		title, _ := webDriver.Title()
		if title == "" || title == "Just a moment..." {
			retry(webDriver)
		}
	}

	return err == nil
}

func retry(webDriver selenium.WebDriver) {
	time.Sleep(time.Second)
	err := webDriver.Refresh()
	if err == nil {
		HandleCaptcha(webDriver)
	}
}
