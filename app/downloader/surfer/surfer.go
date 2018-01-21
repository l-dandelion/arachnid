package surfer

import (
	"net/http"
	"sync"
)

const (
	DOWNLOADERID_SURF = 0
)

type Surfer interface {
	Download(req Request) (resp *http.Response, err error)
}

var (
	surf      Surfer
	once_surf sync.Once
)

func Download(req Request) (resp *http.Response, err error) {
	switch req.GetDownloaderId() {
	case DOWNLOADERID_SURF:
		once_surf.Do(func() {
			surf = &Surf{}
		})
		resp, err = surf.Download(req)
	}
	return resp, err
}
