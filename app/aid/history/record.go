package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Record struct {
	fileName string
	new      map[string]interface{}
	old      map[string]interface{}
	sync.RWMutex
}

func NewRecord(fileName string) *Record {
	record := &Record{
		fileName: fileName,
		new:      make(map[string]interface{}),
		old:      make(map[string]interface{}),
	}
	return record
}

func (self *Record) Init() {
	self.Lock()
	defer self.Unlock()
	self.new = make(map[string]interface{})
	self.old = make(map[string]interface{})
}

//添加一条记录，返回是否添加成功
func (self *Record) Push(key string, value interface{}) bool {
	self.Lock()
	defer self.Unlock()

	if self.new[key] != nil {
		return false
	}

	if self.old[key] != nil {
		return false
	}

	self.new[key] = value
	return true
}

//查询是否有该记录
func (self *Record) HasRecord(key string) bool {
	self.RLock()
	defer self.RUnlock()
	return self.new[key] != nil || self.old[key] != nil
}

//删除某个未被输出的记录
func (self *Record) DeleteNew(key string) bool {
	self.Lock()
	defer self.Unlock()
	if self.new[key] != nil {
		delete(self.new, key)
		return true
	}
	return false
}

//将新纪录输出
func (self *Record) Flush(provider string) (length int, err error) {
	self.Lock()
	defer self.Unlock()
	length = len(self.new)
	fmt.Println(length)
	if length == 0 {
		return
	}
	switch provider {
	default:
		var f *os.File
		f, err = os.OpenFile(self.fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		b, _ := json.Marshal(self.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()
		for key, value := range self.new {
			self.old[key] = value
		}
	}
	self.new = make(map[string]interface{})
	return
}

//读取历史记录
func (self *Record) Read(provider string) (length int, err error) {
	self.Lock()
	defer self.Unlock()

	switch provider {
	default:
		var f *os.File
		f, err = os.OpenFile(self.fileName, os.O_CREATE, 0777)
		if err != nil {
			return 0, nil
		}
		b, _ := ioutil.ReadAll(f)
		f.Close()
		if len(b) == 0 {
			return
		}
		b[0] = '{'
		json.Unmarshal(append(b, '}'), &self.old)
	}

	return len(self.old), err
}

func (self *Record) GetNew() map[string]interface{} {
	self.RLock()
	defer self.RUnlock()
	return self.new
}

func (self *Record) GetOld() map[string]interface{} {
	self.RLock()
	defer self.RUnlock()
	return self.old
}
