package surfer

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Request interface {
	GetUrl() string
	GetProxy() string
	GetMethod() string
	GetHeader() http.Header
	GetPostData() string
	GetDialTimeout() time.Duration
	GetConnTimeout() time.Duration
	GetRetryTimes() int
	GetRetryPauseTime() time.Duration
	GetRedirectTimes() int
	GetCookieJar() *cookiejar.Jar
	GetDownloaderId() int
}
