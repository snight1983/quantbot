package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"path/filepath"
	"strconv"
	"time"
	"unsafe"

	"go.uber.org/zap"
)

// GetWithJSON .
func GetWithJSON(url string, headers map[string]string, cookie *cookiejar.Jar) ([]byte, error) {
	reqest, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return nil, err
	}
	for key, value := range headers {
		reqest.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if cookie != nil {
		client.Jar = cookie
	}
	response, err := client.Do(reqest)
	if err != nil {
		return nil, err
	}
	if nil != response {
		defer response.Body.Close()
		respBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return respBytes, nil
	}
	return nil, errors.New("no body")
}

// DeleteWithJSON .
func DeleteWithJSON(url string, headers map[string]string) (string, error) {
	var result string
	reqest, err := http.NewRequest("DELETE", url, nil)
	if nil != err {
		return result, err
	}

	for key, value := range headers {
		reqest.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(reqest)
	if err != nil {
		return result, err
	}
	if nil != response {
		defer response.Body.Close()
		respBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return result, err
		}
		result = *(*string)(unsafe.Pointer(&respBytes))
	}
	return result, nil
}

// DelateWithJSON .
func DelateWithJSON(url string, headers map[string]string) (string, error) {
	var result string
	reqest, err := http.NewRequest("DELETE", url, nil)
	if nil != err {
		return result, err
	}

	for key, value := range headers {
		reqest.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	response, err := client.Do(reqest)

	if err != nil {
		return result, err
	}
	if nil != response {
		defer response.Body.Close()
		respBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return result, err
		}
		result = *(*string)(unsafe.Pointer(&respBytes))
	}
	return result, nil
}

func HeaderSniff(url string, headers map[string]string) bool {
	reqest, err := http.NewRequest("HEAD", url, nil)
	if nil != err {
		return false
	}
	for key, value := range headers {
		reqest.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(netw, addr, time.Second*2)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * 2,
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(reqest)
	if nil != err {
		Logger.Warn("HeadSniff error",
			zap.String("url", url),
			zap.String("err", err.Error()))
		return false
	}
	if response == nil {
		return false
	}
	defer response.Body.Close()
	return true
}

// HttpREQ .
func HttpREQ(ms string, url string, bf []byte, headers map[string]string) ([]byte, map[string][]string, int, error) {
	var reader = bytes.NewReader(bf)
	reqest, err := http.NewRequest(ms, url, reader)
	if nil != err || nil == reqest {
		return nil, nil, 0, err
	}

	for key, value := range headers {
		reqest.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(reqest)
	if err != nil {
		return nil, nil, 500, err
	}

	if response == nil {
		return nil, nil, 500, errors.New("response.body nil")
	}
	defer response.Body.Close()
	resCode := response.StatusCode

	defer response.Body.Close()
	respBytes, err := ioutil.ReadAll(response.Body)
	if nil != err {
		return respBytes, response.Header, resCode, err
	}
	return respBytes, response.Header, resCode, nil
}

// PostWithJSON .
func PostWithJSON(uri string, bt []byte, headers map[string]string, cookie *cookiejar.Jar) ([]byte, int, error) {
	var reader = bytes.NewReader(bt)
	request, err := http.NewRequest("POST", uri, reader)
	if err != nil {
		return nil, 500, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(netw, addr, time.Second*10)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * 3600,
	}

	client := &http.Client{Transport: tr}
	if cookie != nil {
		client.Jar = cookie
	}
	response, err := client.Do(request)

	if err != nil {
		return nil, 500, err
	}
	if response == nil {
		return nil, 500, errors.New("response.body nil")
	}
	defer response.Body.Close()
	resCode := response.StatusCode
	respBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, 500, err
	}
	return respBytes, resCode, nil
}

// PUTWithJSON .
func PUTWithJSON(uri string, body []byte, headers map[string]string) (string, int, error) {
	var result string

	var reader = bytes.NewReader(body)
	request, err := http.NewRequest("PUT", uri, reader)
	if err != nil {
		return result, 500, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	client := http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return result, 500, err
	}
	if response == nil {
		return "", 500, errors.New("response.body nil")
	}
	defer response.Body.Close()
	resCode := response.StatusCode
	respBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, 500, err
	}
	result = *(*string)(unsafe.Pointer(&respBytes))
	return result, resCode, nil
}

func UploadFormDataJsonFile(opt, url, itemname, filename string, buf []byte, headers map[string]string) (string, error) {
	var result string
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "application/json")
	h.Set("Content-Disposition", fmt.Sprintf("form-data; name=%s; filename=%s", itemname, filename))
	pa, err := w.CreatePart(h)
	if nil != err {
		return "", err
	}
	pa.Write(buf)
	_, err = w.CreateFormFile(itemname, filepath.Base(filename))
	if nil != err {
		return "", err
	}
	w.Close()
	req, err := http.NewRequest(opt, url, body)
	if nil != err {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(req)
	if response == nil {
		return "", err
	}
	if response.Body == nil {
		return "", errors.New("response.body nil")
	}
	defer response.Body.Close()

	respBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	result = *(*string)(unsafe.Pointer(&respBytes))
	return result, nil

}

// UploadFormData .
func UploadFormData(opt, url, name, filename string, buf []byte, headers map[string]string, args map[string]string) (string, int, error) {
	var result string
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)
	fileWriter, err := bodyWriter.CreateFormFile(name, filename)

	if nil != err {
		return result, 500, err
	}

	btReader := bytes.NewReader(buf)
	_, err = io.Copy(fileWriter, btReader)
	if err != nil {
		return "", 500, err
	}

	for key, value := range args {
		err := bodyWriter.WriteField(key, value)
		if err != nil {
			return "", 500, err
		}
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	request, err := http.NewRequest(opt, url, bodyBuffer)
	if err != nil {
		return "", 500, err
	}
	request.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	client := http.Client{}
	response, err := client.Do(request)
	if response == nil {
		return "", 500, err
	}
	if response.Body == nil {
		return "", 500, errors.New("response.body nil")
	}
	defer response.Body.Close()

	respBytes, err := ioutil.ReadAll(response.Body)
	resCode := response.StatusCode
	if err != nil {
		return result, resCode, err
	}
	result = *(*string)(unsafe.Pointer(&respBytes))
	return result, resCode, nil
}

func PostWithFormData(method, url string, postData *map[string]string) {
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	for k, v := range *postData {
		w.WriteField(k, v)
	}
	w.Close()
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, _ := http.DefaultClient.Do(req)
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Printf("%s", data)
}

// UploadFormData .
func UploadFormDataNoFile(opt, url string, headers map[string]string, args *map[string]string, cookie **cookiejar.Jar) (string, int, error) {
	//url = "http://10.20.16.227:31428/dex/auth/local?req=b7rna6nc6ieub55vabirbrgki"
	var result string
	bodyBuffer := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuffer)
	//bodyWriter.SetBoundary("--------------------------249090385668906940555687")
	for key, value := range *args {
		bodyWriter.WriteField(key, value)
	}
	bodyWriter.Close()
	request, err := http.NewRequest(opt, url, bodyBuffer)
	if err != nil {
		return "", 500, err
	}

	request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	request.Header.Set("Content-Length", strconv.FormatInt(request.ContentLength, 10))
	request.Header.Set("Host", request.Host)

	//urli := turl.URL{}
	//urlproxy, _ := urli.Parse("http://127.0.0.1:8888")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//Proxy:           http.ProxyURL(urlproxy),
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	*cookie, _ = cookiejar.New(nil)
	client := &http.Client{Transport: tr, Jar: *cookie}

	response, err := client.Do(request)
	if response == nil {
		return "", 500, err
	}
	if response.Body == nil {
		return "", 500, errors.New("response.body nil")
	}
	defer response.Body.Close()

	respBytes, err := ioutil.ReadAll(response.Body)
	resCode := response.StatusCode
	if err != nil {
		return result, resCode, err
	}
	result = *(*string)(unsafe.Pointer(&respBytes))
	return result, resCode, nil
}

// AddURLParam .
func AddURLParam(url *string, key string, value string, first *bool) {
	if key != "" && value != "" {
		if !(*first) {
			*url += "&"
		}
		*first = false
		*url += key
		*url += "="
		*url += value
	}
}

// DownloadSmallFile .
func DownloadSmallFile1(url string) ([]byte, error) {
	response, err := http.Get(url)
	if nil != err {
		return nil, err
	}
	if nil != response.Body {
		defer response.Body.Close()
	} else {
		return nil, nil
	}
	return ioutil.ReadAll(response.Body)
}

// DownloadReader .
func DownloadReader(url string) (io.ReadCloser, int, error) {
	response, err := http.Get(url)
	if nil != err {
		return nil, 500, err
	}
	if nil == response.Body {
		return nil, 500, errors.New("empty body")
	}
	//fmt.Println(response.Status)
	return response.Body, response.StatusCode, nil
}
