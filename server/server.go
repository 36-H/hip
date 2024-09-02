package main

import (
	"fmt"
	"hip/common"
	"net"
	"time"

	"github.com/astaxie/beego/logs"
)


type Server struct{
	ListenAddr string
	sessionManager *SessionManager
}

func NewServer(listenAddr string,sessionManager *SessionManager) *Server{
	server := &Server{
		ListenAddr: listenAddr,
		sessionManager: sessionManager,
	}
	go server.checkOnlineInterval()
	return server
}

func (s *Server)ListenAndServer() error{
	listener, err := net.Listen("tcp",s.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}

func (s *Server)handleConn(conn net.Conn){
	// defer conn.Close()

	handshakeReq := &common.HandshakeReq{}

	err := handshakeReq.Decode(conn)
	if err != nil {
		logs.Error(fmt.Sprintf("decode handshake fail: %v",err))
		return
	}
	logs.Debug(fmt.Sprintf("Server from Client[%s] handshakeReq %+v", handshakeReq.ClientId, handshakeReq))

	_ ,err = s.sessionManager.CreateSession(
		handshakeReq.ClientId,conn)
	if err != nil {
		logs.Error(fmt.Sprintf("init smux session fail: %v",err))
		return
	}

	// defer s.sessionManager.CloseSession(handshakeReq.ClientId)
}

func (s *Server)checkOnlineInterval(){
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for range tick.C{
		s.sessionManager.Range(func(k string,v *Session) bool{
			return  !v.Connection.IsClosed()
		})
	}
}