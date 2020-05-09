package util

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"main/util/logger"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 发送http请求，不是https!
func Get(url string) string {
	//超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.WithFields(logger.Weblog).Errorln("客户端Get请求短信服务失败：", err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.WithFields(logger.Weblog).Errorln("响应结果读取失败：", err)
		}
	}
	return result.String()
}

//发送post请求
func DoHttpPost(apiUrl string, postParam map[string]string) (result string, err error) {
	postValue := url.Values{}
	for key, value := range postParam {
		var content, _, _ = transform.String(simplifiedchinese.GBK.NewEncoder(), value)
		postValue.Set(key, content)
	}
	fmt.Println("<POST>" + apiUrl)
	fmt.Println("post param : " + postValue.Encode())
	response, err := http.Post(apiUrl, "application/x-www-form-urlencoded;charset=gbk", strings.NewReader(postValue.Encode()))
	if err != nil {
		log.WithFields(logger.Weblog).Errorln("客户端Post请求短信服务失败：", err)
		return "", err
	}
	text, err2 := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err2 != nil {
		log.WithFields(logger.Weblog).Errorln("响应结果读取失败：", err)
		return "", err2
	}
	return string(text), nil
}
