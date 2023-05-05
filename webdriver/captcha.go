package webdriver

import (
	"strings"
	"time"

	"github.com/linweiyuan/go-chatgpt-api/api"
	"github.com/linweiyuan/go-chatgpt-api/util/logger"
	"github.com/tebeka/selenium"
)

const (
	checkCaptchaTimeout      = 15
	checkAccessDeniedTimeout = 5
	checkCaptchaInterval     = 1
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
			return true, nil
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
			return true, nil
		}

		element.Click()
		return true, nil
	}, time.Second*checkCaptchaTimeout, time.Second*checkCaptchaInterval)

	if err != nil {
		webDriver.Refresh()
		HandleCaptcha(webDriver)
	} else {
		title, _ := webDriver.Title()
		if title == "" || title == "Just a moment..." {
			webDriver.Refresh()
			HandleCaptcha(webDriver)
		}
	}

	return err == nil
}

func isAccessDenied(webDriver selenium.WebDriver) bool {
	err := webDriver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		element, err := driver.FindElement(selenium.ByClassName, "cf-error-details")
		if err != nil {
			return false, nil
		}

		accessDeniedText, _ := element.Text()
		logger.Error(accessDeniedText)
		return true, nil
	}, time.Second*checkAccessDeniedTimeout, time.Second*checkCaptchaInterval)

	if err != nil {
		return false
	}

	return true
}
