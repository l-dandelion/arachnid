package proxy

import (
	"sort"
	"sync"
	"time"
)

type ProxyForHost struct {
	curIndex  int             //当前下标
	proxys    []string        //代理
	timeDelay []time.Duration //延时
	sync.RWMutex
}

//返回可用的代理数
func (self *ProxyForHost) Len() int {
	return len(self.proxys)
}

//大小，用于排序
func (self *ProxyForHost) Less(i, j int) bool {
	return self.timeDelay[i] < self.timeDelay[j]
}

//交换，用于排序
func (self *ProxyForHost) Swap(i, j int) {
	self.proxys[i], self.proxys[j] = self.proxys[j], self.proxys[i]
	self.timeDelay[i], self.timeDelay[j] = self.timeDelay[j], self.timeDelay[i]
}

//排序
func (self *ProxyForHost) Sort() {
	self.Lock()
	if self.Len() != 0 {
		sort.Sort(self)
	}
	self.Unlock()
}

func (self *ProxyForHost) NextIndex() bool {
	self.Lock()
	defer self.Unlock()
	if self.curIndex+1 < self.Len() {
		self.curIndex++
		return true
	}
	return false
}

func (self *ProxyForHost) GetOne() string {
	self.RLock()
	defer self.RUnlock()
	if self.Len() <= 0 {
		return ""
	}
	if self.curIndex < self.Len() {
		return self.proxys[self.curIndex]
	}
	return self.proxys[0]
}

func (self *ProxyForHost) Init() {
	self.Lock()
	defer self.Unlock()
	self.curIndex = 0
	self.proxys = []string{}
	self.timeDelay = []time.Duration{}
}

func (self *ProxyForHost) AddProxy(proxy string, timeDelay time.Duration) {
	self.Lock()
	defer self.Unlock()
	self.proxys = append(self.proxys, proxy)
	self.timeDelay = append(self.timeDelay, timeDelay)
}
