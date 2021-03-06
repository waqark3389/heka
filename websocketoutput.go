package heka_websockets_output

import (
        "fmt"
        "github.com/mozilla-services/heka/message"
        "github.com/mozilla-services/heka/pipeline"
        "golang.org/x/net/websocket"
        "net/http"
)

type connection struct {
        ws   *websocket.Conn
        send chan *message.Message
}

type WebSocketsOutputConfig struct {
        Address string `toml:"address"`
        Handler string `toml:"handler"`
}

type WebSocketsOutput struct {
        conf        *WebSocketsOutputConfig
        connections map[*connection]struct{}
        register    chan *connection
        unregister  chan *connection
        broadcast   chan *message.Message
}

func (wso *WebSocketsOutput) ConfigStruct() interface{} {
        return &WebSocketsOutputConfig{}
}

func (wso *WebSocketsOutput) Init(config interface{}) error {
        wso.conf = config.(*WebSocketsOutputConfig)
        wso.connections = make(map[*connection]struct{})
        wso.register = make(chan *connection)
        wso.unregister = make(chan *connection)
        wso.broadcast = make(chan *message.Message, 1028)

        // Connections handler
        go func() {
                var conn *connection
                var m *message.Message
                for {
                        select {
                        case conn = <-wso.register:
                                wso.connections[conn] = struct{}{}
                        case conn = <-wso.unregister:
                                delete(wso.connections, conn)
                                close(conn.send)
                        case m = <-wso.broadcast:
                                for conn = range wso.connections {
                                        select {
                                        case conn.send <- m:
                                        default:
                                                delete(wso.connections, conn)
                                                close(conn.send)
                                                go conn.ws.Close()
                                        }
                                }
                        }
                }
        }()

        // Websocket server and connection handler
        http.Handle(wso.conf.Handler, websocket.Handler(func(ws *websocket.Conn) {
                c := &connection{ws, make(chan *message.Message, 1028)}

                wso.register <- c

                defer func() {
                        wso.unregister <- c
                }()

                var err error
                for m := range c.send {
                        if err = websocket.JSON.Send(ws, m); err != nil {
                                fmt.Println("Websocket:", err.Error())
                                break
                        }
                }
        }))

        go func() {
                if err := http.ListenAndServe(wso.conf.Address, nil); err != nil {
                        fmt.Println("Http:", err.Error())
                }
        }()

        return nil
}

func (wso *WebSocketsOutput) Run(or pipeline.OutputRunner, h pipeline.PluginHelper) error {
        for pc := range or.InChan() {
                wso.broadcast <- pc.Message
                pc.Recycle(nil)
        }
        return nil
}

func init() {
        pipeline.RegisterPlugin("WebSocketsOutput", func() interface{} {
                return new(WebSocketsOutput)
        })
}
