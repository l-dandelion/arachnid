package surfer

import (
	"crypto/tls"
	"github.com/l-dandelion/arachnid/utils"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
	"fmt"
)

type Param struct {
	url       *url.URL       //目标URL
	proxy     *url.URL       //代理
	method    string         //请求方法
	body      io.Reader      //请求体
	header    http.Header    //请求头
	cookieJar *cookiejar.Jar //cookie

	dialTimeout time.Duration //创建连接超时时间
	connTimeout time.Duration //连接状态超时时间

	retryTimes     int           //失败重试次数
	retryPauseTime time.Duration //失败重试间隔
	redirectTimes  int           //重定向最大次数

	client *http.Client //客户端
}

//根据Request构建Param
func NewParam(req Request) (param *Param, err error) {
	param = &Param{}

	//将url字符串转化为*url.URL
	param.url, err = utils.UrlEncode(req.GetUrl())
	if err != nil {
		return nil, err
	}

	if req.GetProxy() != "" {
		param.proxy, err = utils.UrlEncode(req.GetProxy())
		if err != nil {
			return nil, err
		}
	}

	//请求头
	param.header = req.GetHeader()
	if param.header == nil {
		param.header = make(http.Header)
	}

	//cookie
	param.cookieJar = req.GetCookieJar()

	//根据请求方法构造body
	method := strings.ToUpper(req.GetMethod())
	switch method {
	case "HEAD", "GET":
		param.method = method
	case "POST":
		param.method = method
		param.header.Add("Content-Type", "application/x-www-form-urlencoded")
		param.body = strings.NewReader(req.GetPostData())
	default:
		param.method = "GET"
	}

	if req.GetDialTimeout() < 0 {
		param.dialTimeout = 0
	}

	param.connTimeout = req.GetConnTimeout()
	param.retryPauseTime = req.GetRetryPauseTime()
	param.retryTimes = req.GetRetryTimes()
	param.redirectTimes = req.GetRedirectTimes()
	return
}

//将原本的Request信息，写会响应中的Request
func (self *Param) writeBack(resp *http.Response) *http.Response {
	if resp == nil {
		resp = new(http.Response)
		resp.Request = new(http.Request)
	} else if resp.Request == nil {
		resp.Request = new(http.Request)
	}
	resp.Request.Header = self.header
	resp.Request.Method = self.method
	resp.Request.Host = self.url.Host
	return resp
}

//检验是否重定向
func (self *Param) checkRedirect(req *http.Request, via []*http.Request) error {
	//无限制
	if self.redirectTimes == 0 {
		return nil
	}
	//不重定向
	if self.redirectTimes < 0 {
		return fmt.Errorf("not allow redirect.")
	}
	//次数上限
	if len(via) >= self.redirectTimes {
		return fmt.Errorf("stopped after %v redirects", self.redirectTimes)
	}
	return nil
}

//根据参数创建client（待改进：client池，重复使用client）
func (self *Param) buildClient() *http.Client {
	client := new(http.Client)
	//cookie
	if self.cookieJar != nil {
		client.Jar = self.cookieJar
	}
	//重定向检测函数
	client.CheckRedirect = self.checkRedirect

	//构造transport
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, self.dialTimeout)
			if err != nil {
				return nil, err
			}
			if self.connTimeout > 0 {
				c.SetDeadline(time.Now().Add(self.connTimeout))
			}
			return c, nil
		},
	}

	//代理
	if self.proxy != nil {
		transport.Proxy = http.ProxyURL(self.proxy)
	}

	//https处理
	if strings.ToLower(self.url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}

	client.Transport = transport
	self.client = client
	return client
}

//根据参数发送请求获取响应
func (self *Param) httpRequest() (resp *http.Response, err error) {
	req, err := http.NewRequest(self.method, self.url.String(), self.body)
	if err != nil {
		return nil, err
	}

	req.Header = self.header


	for i := 0; i <= self.retryTimes; i++ {
		resp, err = self.client.Do(req)
		if err != nil {
			time.Sleep(self.retryPauseTime)
			continue
		}
		break
	}
	return resp, err
}
