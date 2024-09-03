package main

import (
	"fmt"
	"hip/common"
	"io"
	"net"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/xtaci/smux"
)

type Client struct {
	clientId   string
	serverAddr string
}

func NewClient(clientId string, serverAddr string) *Client {
	return &Client{
		clientId:   clientId,
		serverAddr: serverAddr,
	}
}

func (c *Client) Run(){
	for {
		err := c.run()
		if err != nil && err != io.EOF{
			logs.Error(fmt.Sprintf("err: %v\n", err))
		}
		logs.Warn(fmt.Sprintf("reconnect %s",c.serverAddr))
		time.Sleep(common.ReconnetTimeOut)
	}
} 

func (c *Client) run() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	handshakeReq := &common.HandshakeReq{ClientId: c.clientId}
	buf, err := handshakeReq.Encode()
	if err != nil {
		return err

	}
	logs.Debug(fmt.Sprintf("Client[%s] handshakeReq %+v",c.clientId, handshakeReq))

	conn.SetWriteDeadline(time.Now().Add(common.WriteTimeOut))
	_, err = conn.Write(buf)
	conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}

	mux, err := smux.Client(conn, nil)
	if err != nil {
		return err
	}
	defer mux.Close()

	for {
		stream, err := mux.AcceptStream()
		if err != nil {
			return err
		}

		go c.handleSteam(stream)

	}
}

func (c *Client) handleSteam(stream net.Conn) {
	defer stream.Close()

	proxyProtocol := &common.ProxyProtocol{}
	err := proxyProtocol.Decode(stream)
	if err != nil {
		logs.Error(fmt.Sprintf("decode proxyProtocol fail: %v", err))
		return
	}
	logs.Debug(fmt.Sprintf("Client[%s] proxyProtocol %+v",c.clientId,proxyProtocol))

	switch proxyProtocol.InternalProtocol {
	case "tcp":
		localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", proxyProtocol.InternalIp, proxyProtocol.InternalPort))
		if err != nil {
			logs.Error(fmt.Sprintf("connect to local addr fail: %v", err))
			return
		}
		defer localConn.Close()

		go func() {
			defer localConn.Close()
			defer stream.Close()
			io.Copy(localConn, stream)
		}()
		io.Copy(stream, localConn)
	default:
		logs.Warn(fmt.Sprintf("unsupported protocol %s", proxyProtocol.InternalProtocol))
	}
}
