package main

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linweiyuan/go-chatgpt-api/api"
	"github.com/linweiyuan/go-chatgpt-api/util/logger"
	"github.com/linweiyuan/go-chatgpt-api/webdriver"
)

func init() {
	gin.ForceConsoleColor()
}

type SSE struct {
	NewClients    chan chan string
	ClosedClients chan chan string
	TotalClients  map[chan string]bool

	CookiesChannel chan string
}

type ClientChannel chan string

//goland:noinspection GoUnhandledErrorResult
func main() {
	router := gin.Default()

	sse := New()

	router.GET("/", func(c *gin.Context) {
		cookies, err := webdriver.WebDriver.GetCookies()
		if err != nil {
			webdriver.WebDriver.Refresh()
			c.JSON(http.StatusInternalServerError, api.ReturnMessage(err.Error()))
			return
		}

		responseMap := make(map[string]string)
		for _, cookie := range cookies {
			responseMap[cookie.Name] = cookie.Value
		}
		c.JSON(http.StatusOK, responseMap)
	})

	router.GET("/sse", HeadersMiddleware(), sse.handleClientChannels(), func(c *gin.Context) {
		v, ok := c.Get("clientChannel")
		if !ok {
			return
		}

		clientChannel, ok := v.(ClientChannel)
		if !ok {
			return
		}

		go webdriver.SendCookies()

		c.Stream(func(w io.Writer) bool {
			if cookies, ok := <-clientChannel; ok {
				c.SSEvent(" cookies", " "+cookies) // add a space
				return true
			}

			return false
		})
	})

	go func() {
		for {
			cookies := <-webdriver.CookiesChannel
			sse.CookiesChannel <- cookies
		}
	}()

	router.Run(":8080")
}

func New() (sse *SSE) {
	sse = &SSE{
		NewClients:     make(chan chan string),
		ClosedClients:  make(chan chan string),
		TotalClients:   make(map[chan string]bool),
		CookiesChannel: make(chan string),
	}

	go sse.listen()

	return
}

func (sse *SSE) listen() {
	for {
		select {
		case client := <-sse.NewClients:
			sse.TotalClients[client] = true
			logger.Warn("New connection, total: " + strconv.Itoa(len(sse.TotalClients)))

		case client := <-sse.ClosedClients:
			delete(sse.TotalClients, client)
			close(client)
			logger.Error("Remove connection, total: " + strconv.Itoa(len(sse.TotalClients)))

		case cookies := <-sse.CookiesChannel:
			for client := range sse.TotalClients {
				client <- cookies
			}
		}
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func (sse *SSE) handleClientChannels() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientChannel := make(ClientChannel)
		sse.NewClients <- clientChannel

		defer func() {
			sse.ClosedClients <- clientChannel
		}()

		c.Set("clientChannel", clientChannel)
		c.Next()
	}
}
