/*
 The MIT License (MIT)

 Copyright (c) 2015 Bluek404

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

// https://github.com/Bluek404/gohtml
//go:generate gohtml tpl

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const host = ":8080"

func initRouter() http.Handler {
	staticFiles := make(map[string][]byte)
	filepath.Walk("static", func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		byt, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		staticFiles["/"+path] = byt
		return nil
	})
	return &router{staticFiles}
}

type router struct {
	staticFiles map[string][]byte
}

func (router *router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t := time.Now()
	defer log.Println(r.RemoteAddr, r.Method, r.RequestURI, time.Now().Sub(t).String())

	switch r.RequestURI {
	case "/":
		Index(rw, r)
	case "/ws":
		wsMain(rw, r)
	default:
		v, ok := router.staticFiles[r.RequestURI]
		if ok {
			rw.Write(v)
			return
		}
		rw.WriteHeader(404)
	}
}

func main() {
	log.Println("Running on", host)
	err := http.ListenAndServe(host, initRouter())
	if err != nil {
		log.Fatal(err)
	}
}
