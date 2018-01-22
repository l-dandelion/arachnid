package request

import (
	"net/http"
	"net/http/cookiejar"
	"time"
	"crypto/md5"
	"encoding/hex"
)

type Request struct {
	Spider string //规则名
	Rule   string //规则节点名

	URL          string         //目标URL
	Method       string         //Get/Post/Post-M
	Header       http.Header    //请求头
	PostData     string         //post 数据
	EnableCookie bool           //是否使用cookie
	CookieJar    *cookiejar.Jar //cookie

	DialTimeout   time.Duration //创建超时时间
	ConnTimeout   time.Duration //连接超时时间
	RedirectTimes int           //重定向最大次数

	RetryTimes     int           //重试次数
	RetryPauseTime time.Duration //重试时间间隔

	Priority   int  //优先级
	Reloadable bool //是否允许重复下载

	DownLoaderId int //下载器类型

	proxy  string //代理
	unique string //唯一标识
}

func (self *Request) GetUrl() string {
	return self.URL
}

func (self *Request) GetProxy() string {
	return self.proxy
}

func (self *Request) GetMethod() string {
	return self.Method
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) GetPostData() string {
	return self.PostData
}

func (self *Request) GetDialTimeout() time.Duration {
	return self.DialTimeout
}

func (self *Request) GetConnTimeout() time.Duration {
	return self.ConnTimeout
}

func (self *Request) GetRetryTimes() int {
	return self.RetryTimes
}

func (self *Request) GetRetryPauseTime() time.Duration {
	return self.RetryPauseTime
}

func (self *Request) GetRedirectTimes() int {
	return self.RedirectTimes
}

func (self *Request) GetCookieJar() *cookiejar.Jar {
	return self.CookieJar
}

func (self *Request) GetDownloaderId() int {
	return self.DownLoaderId
}

func (self *Request) GetUnique() string {
	if self.unique == "" {
		block := md5.Sum([]byte(self.Spider + self.Rule + self.URL + self.Method + self.PostData))
		self.unique = hex.EncodeToString(block[:])
	}
	return self.unique
}