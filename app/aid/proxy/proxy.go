package proxy

import (
	"github.com/l-dandelion/arachnid/app/downloader/surfer"
	"github.com/l-dandelion/arachnid/app/request"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	ipRegexp    *regexp.Regexp
	proxyRegexp *regexp.Regexp
	allIps      map[string]string
	all         map[string]bool
	proxys      []string
	onlineCount int64
	usable      map[string]*ProxyForHost
	ticker      *time.Ticker
	tickMinute  int64
	threadPool  chan bool
	surf        surfer.Surfer
	sync.Mutex
}

const (
	CONN_TIMEOUT   = 4 * time.Second
	DIAL_TIMEOUT   = 4 * time.Second
	TRY_TIMES      = 3
	MAX_THREAD_NUM = 1000
	TICKET_MINUTE = 1
)

var ProxyConfigFileName = "proxy.conf"

func New() *Proxy {
	p := &Proxy{
		ipRegexp:    regexp.MustCompile(`[0-9]+\.[0-9]\.[0-9]+\.[0-9]+`),
		proxyRegexp: regexp.MustCompile(`http[s]?://[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+`),
		allIps:      map[string]string{},
		all:         map[string]bool{},
		proxys:      []string{},
		usable:      make(map[string]*ProxyForHost),
		threadPool:  make(chan bool, MAX_THREAD_NUM),
		surf:        surfer.New(),
	}
	p.Update()
	return p
}

func (self *Proxy) Update() *Proxy {

	f, err := os.Open(ProxyConfigFileName)
	if err != nil {
		return self
	}
	b, _ := ioutil.ReadAll(f)
	f.Close()
	self.proxys = self.proxyRegexp.FindAllString(string(b), -1)
	for _, proxy := range self.proxys {
		self.allIps[proxy] = self.ipRegexp.FindString(proxy)
		self.all[proxy] = false
	}

	self.findOnline()

	self.UpdateTicker(TICKET_MINUTE)

	return self
}

func (self *Proxy) findOnline() *Proxy {
	self.onlineCount = 0
	for _, proxy := range self.proxys {
		self.threadPool <- true
		go func(proxy string) {
			//alive, _, _ := ping.Ping(self.allIps[proxy], CONN_TIMEOUT)
			alive := true
			self.Lock()
			self.all[proxy] = alive
			self.onlineCount++
			self.Unlock()

			<-self.threadPool
		}(proxy)
	}
	for len(self.threadPool) > 0 {
		time.Sleep(1 * time.Second)
	}
	return self
}

func (self *Proxy) GetOne(u string) (proxy string) {
	self.Lock()
	defer self.Unlock()
	if self.onlineCount == 0 {
		return
	}

	u2, _ := url.Parse(u)
	if u2.Host == "" {
		return
	}
	var key = u2.Host
	if strings.Count(key, ".") > 1 {
		key = key[strings.Index(key, ".")+1:]
	}

	var proxyForHost = self.usable[key]
	if proxyForHost == nil {
		proxyForHost = new(ProxyForHost)
		self.usable[key] = proxyForHost
		proxyForHost.Init()
		self.testAndSort(key, u2.Scheme+"://"+u2.Host)
	}

	select {
	case <-self.ticker.C:
		if !proxyForHost.NextIndex() {
			proxyForHost = self.testAndSort(key, u2.Scheme+"://"+u2.Host)
		}
	default:
	}

	proxy = proxyForHost.GetOne()
	return
}

func (self *Proxy) testAndSort(key string, testHost string) (proxyForHost *ProxyForHost) {
	proxyForHost = self.usable[key]
	proxyForHost.Init()
	for proxy, online := range self.all {
		if !online {
			continue
		}
		self.threadPool <- true
		go func(proxy string) {
			alive, timeDelay := self.findUsable(proxy, testHost)
			if alive {
				proxyForHost.AddProxy(proxy, timeDelay)
			}
			<-self.threadPool
		}(proxy)
	}
	for len(self.threadPool) > 0 {
		time.Sleep(0.2e9)
	}
	proxyForHost.Sort()
	return
}

func (self *Proxy) findUsable(proxy, testHost string) (alive bool, timeDelay time.Duration) {
	t0 := time.Now()
	req := &request.Request{
		URL:         testHost,
		Method:      "HEAD",
		ConnTimeout: CONN_TIMEOUT,
		DialTimeout: DIAL_TIMEOUT,
		RetryTimes:  TRY_TIMES,
		Header:      make(http.Header),
	}
	req.SetProxy(proxy)
	_, err := self.surf.Download(req)
	return err == nil, time.Since(t0)
}

func (self *Proxy) UpdateTicker(tickMinute int64) {
	self.Lock()
	defer self.Unlock()
	self.tickMinute = tickMinute
	self.ticker = time.NewTicker(time.Duration(self.tickMinute) * time.Minute)
	for _, proxyForHost := range self.usable {
		proxyForHost.NextIndex()
	}
}
