package dls_cli

import (
	"net"
	"github.com/fatih/pool"
	"github.com/gorilla/websocket"
	"fmt"
	"net/url"
	"encoding/json"
	"errors"
	"log"
	"runtime/debug"
	"sync"
)

var socket_pool pool.Pool
var ws_url *url.URL
var mapLock sync.RWMutex
var conMap map[int]*websocket.Conn = make(map[int]*websocket.Conn)

func cache_ws_conn(c net.Conn, ws *websocket.Conn)  {
	mapLock.Lock()
	defer mapLock.Unlock()
	conMap[c.LocalAddr().(*net.TCPAddr).Port] = ws
}

func get_ws_conn(c net.Conn) *websocket.Conn{
	mapLock.RLock()
	defer  mapLock.RUnlock()
	return conMap[c.LocalAddr().(*net.TCPAddr).Port]
}

func CreatePool(host string, initSize int, maxSize int)  error{
	u, err0 := url.Parse(fmt.Sprintf("ws://%s/websocket", host))
	ws_url = u
	if err0 != nil{
		return err0
	}
	factory := func() (net.Conn, error) {
		ret, err1 := net.Dial("tcp", host)
		if err1 != nil{
			return nil, err1
		}
		ws, _, err2 :=  websocket.NewClient(ret, ws_url, nil, 1024, 1024)
		if err2 != nil{
			ret.Close()
			return nil, err2
		}
		cache_ws_conn(ret, ws)
		return ret, nil
	}
	p, err := pool.NewChannelPool(initSize, maxSize, factory)
	if err != nil{
		return err
	}
	socket_pool = p
	return nil
}

func GetPool()  pool.Pool{
	return socket_pool
}

func SyncDo(key string, action func())  error{
	c, err := socket_pool.Get()
	if err != nil{
		return err
	}
	ws_conn := get_ws_conn(c)
	if ws_conn == nil{
		return errors.New("cant believe that")
	}
	lock_msg, _ := json.Marshal(map[string]string{"cmd": "lock", "key" : key})
	err2 := ws_conn.WriteMessage(websocket.TextMessage, lock_msg)
	if err2 != nil{
		return err2
	}
	_, bytes, err3 := ws_conn.ReadMessage()
	if err3 != nil{
		return err3
	}
	var m1 map[string]string
	err4 := json.Unmarshal(bytes, &m1)
	if err4 != nil{
		return err4
	}
	if m1["msg"] != "get_lock" || m1["key"] != key{
		return errors.New("unrecoginzed msg")
	}
	func() {
		defer func() {
			if panicErr := recover(); panicErr != nil{
				log.Println(panicErr)
				debug.PrintStack()
			}
		}()
		action()
	}()
	unlock_msg, _ := json.Marshal(map[string]string{"cmd": "unlock", "key" : key})
	err5 := ws_conn.WriteMessage(websocket.TextMessage, unlock_msg)
	if err5 != nil{
		return err5
	}
	_, bytes1, err6 := ws_conn.ReadMessage()
	if err6 != nil{
		return err6
	}
	var m2 map[string]string
	err7 := json.Unmarshal(bytes1, &m2)
	if err7 != nil{
		return err7
	}
	if m2["msg"] != "unlock" || m2["key"] != key{
		return errors.New("unrecoginzed msg")
	}
	c.Close()
	return nil
}

func StopPool()  {
	socket_pool.Close()
}