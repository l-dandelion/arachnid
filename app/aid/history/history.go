package history

import (
	"fmt"
	"sync"
)

type Historier interface {
	ReadSuccess(provider string) (length int, err error)
	PushSuccess(req Request) bool
	HasSuccess(req Request) bool
	DeleteSuccess(req Request) bool
	FlushSuccess() (length int, err error)

	ReadFailure(provider string) (length int, err error)
	PushFailure(req Request) bool
	DeleteFailure(req Request) bool
	FlushFailure() (length int, err error)
	HasFailure(req Request) bool
	//测试用
	Output()
}

type History struct {
	success  *Record
	failure  *Record
	provider string
	sync.RWMutex
}

const (
	SUCCESS_SUFFIX = "__y"
	FAILURE_SUFFIX = "__n"
	SUCCESS_FILE   = "./history" + SUCCESS_SUFFIX
	FAILURE_FILE   = "./history" + FAILURE_SUFFIX
)

func New(name string, subName string) Historier {
	successFileName := SUCCESS_FILE + "__" + name
	failureFileName := FAILURE_FILE + "__" + name
	if subName != "" {
		successFileName += "__" + subName
		failureFileName += "__" + subName
	}

	return &History{
		success: NewRecord(successFileName),
		failure: NewRecord(failureFileName),
	}

}

//读取成功的历史记录
func (self *History) ReadSuccess(provider string) (length int, err error) {
	self.Lock()
	self.provider = provider
	self.Unlock()
	return self.success.Read(provider)
}

//添加成功记录
func (self *History) PushSuccess(req Request) bool {
	return self.success.Push(req.GetUnique(), req)
}

//查找成功记录
func (self *History) HasSuccess(req Request) bool {
	key := req.GetUnique()
	return self.success.HasRecord(key)
}

//删除成功记录
func (self *History) DeleteSuccess(req Request) bool {
	return self.success.DeleteNew(req.GetUnique())
}

//输出成功记录
func (self *History) FlushSuccess() (length int, err error) {
	self.RLock()
	provider := self.provider
	self.RUnlock()
	length, err = self.success.Flush(provider)
	return
}

//读取失败历史记录
func (self *History) ReadFailure(provider string) (length int, err error) {
	self.Lock()
	self.provider = provider
	self.Unlock()

	return self.failure.Read(provider)
}

//添加失败历史记录
func (self *History) PushFailure(req Request) bool {
	return self.failure.Push(req.GetUnique(), req)
}

//删除失败历史记录
func (self *History) DeleteFailure(req Request) bool {
	return self.failure.DeleteNew(req.GetUnique())
}

//输出失败历史记录
func (self *History) FlushFailure() (length int, err error) {
	self.RLock()
	provider := self.provider
	defer self.RUnlock()
	return self.failure.Flush(provider)
}

//检查是否有失败历史记录
func (self *History) HasFailure(req Request) bool {
	key := req.GetUnique()
	return self.failure.HasRecord(key)
}

func (self *History) Output() {
	fmt.Println(self.success.new)
	fmt.Println(self.success.old)
	fmt.Println("------------------")
	fmt.Println(self.failure.new)
	fmt.Println(self.failure.old)
}
