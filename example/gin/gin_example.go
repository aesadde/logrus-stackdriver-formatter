package main

import (
	"fmt"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	stackdriver "github.com/aesadde/logrus-stackdriver-formatter"
	"github.com/gin-gonic/gin"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/sirupsen/logrus"
)

func main() {
	//Setup StackDriver logger
	logger := logrus.New()

	formatter := stackdriver.NewFormatter(
		stackdriver.WithService("Example"),
		stackdriver.WithVersion("0.0.1"),
		stackdriver.WithProjectID("ak-app-01"),
	)
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	hook := logrus_stack.NewHook(logrus.AllLevels, logrus.AllLevels)
	logger.AddHook(hook)
	logger.SetReportCaller(true)

	//Create router and add default middleware
	//Recovery handles panics and 500 errors
	router := gin.New()
	router.Use(stackdriver.GinLogger(logger), gin.Recovery())

	//Register common endpoints
	router.GET("/ping", func(c *gin.Context) {
		u, _ := uuid.NewV4()
		c.Header("X-Request-ID", u.String())
		c.JSON(200, "pong")
	})

	router.GET("/test-err", func(c *gin.Context) {
		u, _ := uuid.NewV4()
		c.Header("X-Request-ID", u.String())
		c.Error(fmt.Errorf("This is a test error - Should be logged"))
		c.JSON(500, "pong")
	})

	router.Run(":5000")
}
