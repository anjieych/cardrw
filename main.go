package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/karalabe/hid"
	"log"
	"net/http"
	"sync"
)

func main() {
	port := flag.Int("p", 18999, "指定HTTP端口")
	flag.Parse()
	var deviceLock sync.Mutex

	var device *hid.Device
	read := func() (data string, err error) {
		deviceLock.Lock()
		defer deviceLock.Unlock()
		if device == nil {
			if device, err = deviceFind(); err != nil {
				return
			} else {
				if h, s, f, err := CmdInit(device); err != nil {
					return data, err
				} else {
					fmt.Printf("Device info is hard:%d\tsoft:%d\t,interface:%d\t(1:COM,2:USB)\n", h, s, f)
				}
			}
		}

		if d, err := CmdReadNO_Mifare(device); err == nil {
			data = string(d)
			return data, err
		} else {
			//if read error then redo init device
			device = nil
		}
		return
	}

	http.HandleFunc("/anjieych/cardrw/read", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "text/html; charset=utf-8")

		result := make(map[string]interface{})

		if data, err := read(); err != nil {
			result["result"] = "1001"
			result["message"] = err.Error()
			result["data"] = data
		} else {
			result["result"] = "0"
			result["message"] = "success"
			result["data"] = data
		}
		buf, _ := json.Marshal(result)
		w.Write(buf)
	})
	fmt.Println("http listen : 0.0.0.0: ", *port)
	if err := http.ListenAndServe(":"+fmt.Sprint(*port), nil); err != nil {
		log.Fatal(err)
	}

}
