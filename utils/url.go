package utils

import "net/url"

func UrlEncode(urlStr string)(*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	if err == nil {
		urlObj.RawQuery = urlObj.Query().Encode()
	}
	return urlObj, err
}