package stackdriver

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gin-gonic/gin"
	"github.com/hellofresh/logging-go/context"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware is a middleware for writing request logs in a stuctured
// format to stackdriver.
func LoggingMiddleware(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.New(r.Context()))

			// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
			request := &HTTPRequest{
				RequestMethod: r.Method,
				RequestURL:    r.RequestURI,
				RemoteIP:      r.RemoteAddr,
				Referer:       r.Referer(),
				UserAgent:     r.UserAgent(),
				RequestSize:   strconv.FormatInt(r.ContentLength, 10),
			}

			m := httpsnoop.CaptureMetrics(handler, w, r)

			request.Status = strconv.Itoa(m.Code)
			request.Latency = fmt.Sprintf("%.9fs", m.Duration.Seconds())
			request.ResponseSize = strconv.FormatInt(m.Written, 10)

			fields := logrus.Fields{"httpRequest": request}

			// No idea if this works
			traceHeader := r.Header.Get("X-Cloud-Trace-Context")
			if traceHeader != "" {
				fields["trace"] = traceHeader
			}

			log.WithFields(fields).Info("Completed request")
		})
	}
}

// GinLogger is the logrus logger handler
func GinLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		// other handler can change c.Path so:
		path := c.Request.URL.Path
		method := c.Request.Method
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		requestS := c.Request.ContentLength
		responseS := c.Request.Response.ContentLength
		if requestS < 0 {
			requestS = 0
		}
		request := &HTTPRequest{
			RequestMethod: method,
			RequestURL:    path,
			RemoteIP:      clientIP,
			Referer:       referer,
			UserAgent:     clientUserAgent,
			ResponseSize:  strconv.FormatInt(responseS, 10),
			Latency:       strconv.Itoa(latency),
			Status:        strconv.Itoa(statusCode),
			RequestSize:   strconv.FormatInt(requestS, 10),
		}

		fields := logrus.Fields{"httpRequest": request}

		traceHeader := c.GetHeader("X-Request-ID")
		if traceHeader != "" {
			fields["trace"] = traceHeader
		}

		log.WithFields(fields).Info("Completed request")
	}
}
