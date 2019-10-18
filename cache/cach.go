package cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Response struct {
	body []byte
	headers map[string][]string
	cookie []*http.Cookie
	statusCode int
}

type Request struct {
	path string
	body []byte
	params map[string][]string
	method string
	headers map[string][]string
	cookies []*http.Cookie
	port int
}

func (r *Request) ReadFrom(request *http.Request) error {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}

	r.path = request.URL.Path
	r.params = request.URL.Query()
	r.method = request.Method
	r.headers = request.Header
	r.cookies = request.Cookies()
	r.port = 11223
	r.body = body

	return nil
}

func SendHttpRequest(request Request) (response Response, err error) {

	httpRequest, err := http.NewRequest(request.method, request.Url(), bytes.NewReader(request.body))

	httpRequest.Header = request.headers

	for _, cookie := range request.cookies {
		httpRequest.AddCookie(cookie)
	}

	if err != nil {
		return
	}

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return
	}
	defer httpResponse.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(httpResponse.Body); err != nil {
		return
	}

	response.headers = httpResponse.Header
	response.cookie = httpResponse.Cookies()
	response.statusCode = httpResponse.StatusCode

	html := string(body)
	response.body = []byte(html)

	return
}

func Add(r Request, data Response) {
	if r.method != http.MethodGet  || data.statusCode != 200 {
		return
	}
	found := false
	for header, value := range data.headers {
		if header == "Content-Type" {
			for _, v := range value {
				if strings.Contains(v, "text/html") {
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}
	if !found {
		return
	}
	item := Bucket{
		request: r,
		response: data,
	}
	log.Println("add to cache")
	cache[r.path] = &item
}

func (r Response) Proxy(w http.ResponseWriter) error {

	for k, v := range r.headers {
		w.Header().Set(k, v[0])
	}

	for _, cookie := range r.cookie {
		http.SetCookie(w, cookie)
	}

	w.WriteHeader(r.statusCode)
	_, err := w.Write(r.body)
	return err
}

func Get(request Request) (response Response, found bool) {

	if request.method != http.MethodGet {
		return
	}
	item := cache[request.path]
	if item == nil {
		return
	}
	response = item.response
	found = true
	return
}

func (r Request) Url() (url string) {
	var query string
	for k, v := range r.params {
		query = fmt.Sprintf("%v&%v=%v", query, k, v[0])
	}
	url = fmt.Sprintf("http://%v:%v%v?%v", "localhost", r.port, r.path, query)
	return
}

func (r Response) Read() (headers map[string][]string, cookie []*http.Cookie, statusCode int, body []byte) {
	headers = r.headers
	cookie = r.cookie
	statusCode = r.statusCode
	body = r.body
	return
}

type Bucket struct {
	request Request
	response Response
}

var cache map[string]*Bucket

func init()  {
	cache = make(map[string]*Bucket, 1000)
}