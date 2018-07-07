package main

import (
	"errors"
	"fmt"
	"github.com/karalabe/hid"
	"time"
	"encoding/json"
)
// Cmdrw command to read or write
type Cmdrw struct {
	Head byte
	Len  byte
	Cmd  byte
	Pn   []byte
	Sum  byte
}
// getBytes get []byte of Cmdrw c
func (c *Cmdrw) getBytes() (cmd []byte) {
	fmt.Println(c)
	c.Sum = c.Cmd
	for _, v := range c.Pn {
		c.Sum += v
	}
	c.Len = byte(1 + len(c.Pn)+1 )
	cmd = append(cmd, c.Head, c.Len, c.Cmd)
	cmd = append(cmd, c.Pn...)
	cmd = append(cmd, c.Sum)
	fmt.Println(cmd)
	// 转换成小端模式 ?
	//reverse(cmd)
	//fmt.Println(cmd)
	return

}

// deviceFind find device
func deviceFind()(device *hid.Device,err error){
	devices := hid.Enumerate(uint16(0), 0)
	bFinded:=false
	var deviceInfo *hid.DeviceInfo
	for i, v := range devices {
		data, _ := json.Marshal(v)
		fmt.Println(i, ":", string(data))
		if v.VendorID == 1155 && v.ProductID == 22352 {
			fmt.Println(i, ":", string(data))
			deviceInfo = &v
			bFinded = true
			break
		}
	}
	if bFinded == true && deviceInfo.VendorID == 1155 {
		device, err = deviceInfo.Open()
		if err != nil {
			return
		}

	} else {
		fmt.Println("Cann't find device.")
	}
	return
}

func deviceReadWithTimeout(d *hid.Device) ([]byte, error) {
	buf := make([]byte, 64, 64)
	done := make(chan error)
	go func() {
		_, err := d.Read(buf)
		done <- err
	}()

	timeout := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second)
		timeout <- true
	}()

	select {
	case err := <-done:
		return buf, err
	case <-timeout:
		return nil, errors.New("timed out")
	}

	return nil, errors.New("can't get here")

}
// CmdInit init card reader/writer
func CmdInit(d *hid.Device)(hard byte,soft byte,iface byte,err error) {
	cmdInit := &Cmdrw{
		Head: 0x13,
		Len:  0x04,
		Cmd:  byte('I'),
		Pn:   []byte{byte(0), byte('A')},
	}

	n, err := d.Write(cmdInit.getBytes())
	if err != nil {
		return
	} else {
		fmt.Printf("write to device :%d bytes\n", n)
	}
	b, err := deviceReadWithTimeout(d)
	//b:=make([]byte,64)
	//n,err=d.Read(b)
	if err != nil {
		return
	} else {
		fmt.Println("read: ", b)
	}
	// check read buf
	if b[0]==0x13&&b[2]==byte('I'){
		hard=b[3]
		soft=b[4]
		iface=b[5]
		err=nil
		return
	}else {
		err=errors.New("invalid response")
		return
	}
}

func CmdReadNO_Mifare(d *hid.Device)(string,error){
	cmdRead := &Cmdrw{
		Head: 0X13,
		Len:  0x04,
		Cmd:  byte('S'),
		Pn:   []byte{byte('A'), byte('M')},
	}
	buf_w:=cmdRead.getBytes()
	n, err := d.Write(buf_w)
	if err != nil {
		return "",err
	} else {
		fmt.Println("write to device :",n,buf_w )
	}
	buf_r, err := deviceReadWithTimeout(d)
	//buf_r:=make([]byte,64)
	//n,err=d.Read(buf_r)
	if err != nil {
		return "",err
	} else {
		fmt.Println("read: ",buf_r)
	}
	// check read buf
	if buf_r[0]!=0x13||buf_r[2]!=byte('S')||buf_r[3]!=byte('M'){
		return "",errors.New("invalid response")
	}else {
		// get card_no
		return fmt.Sprintf("%X%X%X%X",buf_r[4],buf_r[5],buf_r[6],buf_r[7]),nil
	}
}