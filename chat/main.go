package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"os"

	"github.com/kazuhei/goblueprints/trace"
)

type templateHandler struct {
	Once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.Once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, req)
}

func main() {
	addr := flag.String("addr", ":1234", "アプリケーションのアドレス")
	flag.Parse()
	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	go r.run()
	log.Println("Webサーバーを起動します。ポート: ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("エラー")
	}
}
