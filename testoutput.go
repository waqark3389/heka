package testoutput

import (
	"io"
	"net/http"
	. "github.com/mozilla-services/heka/pipeline"
	"github.com/mozilla-services/heka/plugins"
	"golang.org/x/net/websocket"
)

func echoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func init() {
	RegisterPlugin("TestOutput", func() interface{} {
		return new(TestOutput)
	})
}
