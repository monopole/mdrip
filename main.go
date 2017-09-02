package main

import (
	"log"
	"os"

	"time"

	"github.com/gorilla/websocket"
	"github.com/monopole/mdrip/config"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/subshell"
	"github.com/monopole/mdrip/webserver"
)

func closeSocket(c *websocket.Conn, done chan struct{}) {
	defer c.Close()
	// Send a close frame, wait for the other side to close the connection.
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func runProxy(addr string) {
	done := make(chan struct{})

	log.Printf("connecting to %s", addr)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer closeSocket(c, done)

	messages := make(chan string)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			messages <- string(message)
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, []byte("hey there"))
	if err != nil {
		log.Println("trouble saying hello:", err)
		return
	}

	for {
		select {
		case m := <-messages:
			log.Printf("recv: %s", m)
			// TODO: Cancel previous timeout, start new one ??
		case <-done:
			return
		case <-time.After(10 * time.Minute):
			// *Always* stop after this super-timeout.
			return
		}
	}
}

func main() {
	c := config.GetConfig()
	p := program.NewProgram(c.ScriptName(), c.FileNames())

	switch c.Mode() {
	case config.ModeTmuxProxy:
		runProxy(string(c.FileNames()[0]))
	case config.ModeServer:
		webserver.NewWebserver(p).Serve(c.HostAndPort())
	case config.ModeTest:
		p.Reload()
		s := subshell.NewSubshell(c.BlockTimeOut(), p)
		if r := s.Run(); r.Problem() != nil {
			r.Print(c.ScriptName())
			if !c.IgnoreTestFailure() {
				log.Fatal(r.Problem())
			}
		}
	default:
		p.Reload()
		if c.Preambled() > 0 {
			p.PrintPreambled(os.Stdout, c.Preambled())
		} else {
			p.PrintNormal(os.Stdout)
		}
	}
}
