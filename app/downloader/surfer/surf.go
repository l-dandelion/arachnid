package surfer

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
)

type Surf struct{}

func New() Surfer {
	return &Surf{}
}

//根据请求下载并返回响应
func (self *Surf) Download(req Request) (resp *http.Response, err error) {
	//规范化参数
	param, err := NewParam(req)
	if err != nil {
		return nil, err
	}
	//创建客户端
	param.buildClient()
	//发送请求
	resp, err = param.httpRequest()
	if err != nil {
		return nil, err
	}
	//根据编码解编
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(resp.Body)
		if err == nil {
			resp.Body = gzipReader
		}
	case "deflate":
		resp.Body = flate.NewReader(resp.Body)
	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(resp.Body)
		if err == nil {
			resp.Body = readCloser
		}
	}
	//将请求信息回写响应
	resp = param.writeBack(resp)

	return resp, err
}
