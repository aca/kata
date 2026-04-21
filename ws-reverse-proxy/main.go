package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
)

// struct implementing http.Handler
type WS struct {
	upgrader *websocket.Upgrader
	dst      string
}

func (ws *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	upgrade := false

	for _, header := range r.Header["Upgrade"] {
		if header == "websocket" {
			upgrade = true
			break
		}
	}

	if upgrade == false {
		remote, err := url.Parse("http://localhost:9222")
		if err != nil {
			panic(err)
		}

		handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				log.Println(r.URL)
				r.Host = remote.Host
				w.Header().Set("X-Ben", "Rad")
				p.ServeHTTP(w, r)
			}
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		handler(proxy)(w, r)

		//       proxy := httputil.NewSingleHostReverseProxy(remote)
		// rp := &httputil.ReverseProxy{
		// 	Director: func(r *http.Request) {
		// 		u, err := url.Parse("http://localhost:9222")
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		r.Host = u.Host
		// 	},
		// }
		// log.Println("reverse proxy")
		// rp.ServeHTTP(w, r)
		return
	}

	// log.Println("header", spew.Sdump(r.Header))
	r.Header.Set("Connection", "Upgrade")
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
		return
	}

	remote, err := url.Parse("ws://localhost:9222")
	if err != nil {
		panic(err)
	}

	r.URL.Host = remote.Host
	r.URL.Scheme = remote.Scheme
	log.Println(r.URL.String())

	header := r.Header.Clone()
	header.Del("Sec-Websocket-Version")
	header.Del("Sec-Websocket-Extensions")
	header.Del("Sec-Websocket-Key")
	// header.Set("Connection", "Upgrade")
	header.Del("Connection")
	header.Del("Upgrade")

	log.Println("header", spew.Sdump(r.Header))


	// dial for communicating to the destination
	dial, _, err := websocket.DefaultDialer.Dial(r.URL.String(), header)
	if err != nil {
		panic(err)
		// w.Write([]byte(err.Error()))
	}
	defer dial.Close()

	for {
		// this data is from the user
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("message type", messageType)
			log.Println("read", string(data))
			log.Println("error", err)
			// panic(err)
		}

		log.Println("message type", messageType)
		log.Println("read", string(data))

		// writes the user's message to the destination
		err = dial.WriteMessage(messageType, data)
		if err != nil {
			panic(err)
		}

		// receives from the destination
		_, dstReader, err := dial.NextReader()
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			panic(err)
		}

		// read the destination's message
		dstData, err := io.ReadAll(dstReader)
		if err != nil {
			panic(err)
		}

		// writes the destination's message to the user
		err = conn.WriteMessage(messageType, dstData)
		if err != nil {
			panic(err)
		}
		log.Println("write", string(dstData))
	}
}

func main() {
	// logger setting

	// url for proxying
	// httpDst := "http://127.0.0.1:8545"
	wsDst := "ws://127.0.0.1:9222"
	// u, err := url.Parse(httpDst)
	// if err != nil {
	// 	logger.Fatal().Err(err).Send()
	// }

	// channel to receive a routing error

	// // http route (port 3000)
	// go func() {
	// 	rp := &httputil.ReverseProxy{
	// 		Director: func(r *http.Request) {
	// 			logger.Info().Any("request", map[string]any{
	// 				"path":   r.URL.Path,
	// 				"method": r.Method,
	// 				"ip":     r.RemoteAddr,
	// 			}).Send()
	// 			r.URL = u
	// 		},
	// 	}
	// 	http.Handle("/http", rp)
	// 	if err := http.ListenAndServe(":3000", nil); err != nil {
	// 		ch <- err
	// 	}
	// }()

	// websocket route (port 3001)
	ws := &WS{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		dst: wsDst,
	}
	http.Handle("/", ws)
	if err := http.ListenAndServe(":3001", nil); err != nil {
		panic(err)
	}

}
