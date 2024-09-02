package main

import (
	"fmt"
	"hip/common"
	"io"
	"net"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

type Listener struct {
	proxyProtocol  *common.ProxyProtocol
	inner          net.Listener
	sessionManager *SessionManager
	CloseOnce sync.Once
	close     chan struct{}
}

func NewListener(proxyProtocol *common.ProxyProtocol,sessionManager *SessionManager) *Listener {
	return &Listener{
		proxyProtocol: proxyProtocol,
		close:          make(chan struct{}),
		sessionManager: sessionManager,
	}
}

func (l *Listener) ListenAndServerTCP() error {
	listener, err := net.Listen("tcp",
		fmt.Sprintf("%s:%d", l.proxyProtocol.PublicIP, l.proxyProtocol.PublicPort))
	if err != nil {
		return err
	}
	defer listener.Close()
	l.inner = listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go l.handleTcp(conn)
	}
}

func (l *Listener) ListenAndServer() error {
	switch l.proxyProtocol.PublicProtocol {
	case "tcp":
		return l.ListenAndServerTCP()
	default:
		return fmt.Errorf("TODO://")
	}
}

func (l *Listener) handleTcp(conn net.Conn) {
	defer conn.Close()

	//查询session
	tunelConnet, err := l.sessionManager.GetSesssionByClientId(l.proxyProtocol.ClientId)
	if err != nil {
		logs.Error(fmt.Sprintf("Get session for client %s fail", l.proxyProtocol.ClientId))
		return
	}
	defer tunelConnet.Close()
	//封装proxyprotocol
	pp, err := l.proxyProtocol.Encode()
	if err != nil {
		logs.Error(fmt.Sprintf("encode proxyProtocol fail: %v", err))
		return
	}
	logs.Debug(fmt.Sprintf("Server to Client[%s] proxyProtocol %+v", l.proxyProtocol.ClientId, l.proxyProtocol.ClientId))

	tunelConnet.SetWriteDeadline(time.Now().Add(common.WriteTimeOut))
	_, err = tunelConnet.Write(pp)
	tunelConnet.SetDeadline(time.Time{})
	if err != nil {
		logs.Error(fmt.Sprintf("write proxyProtocol fail: %v", err))
		return
	}

	//数据拷贝
	go func ()  {
		defer tunelConnet.Close()
		defer conn.Close()
		io.Copy(tunelConnet, conn)	
	}()
	//这里会阻塞 所以上面协程的代码不能放到下方去
	io.Copy(conn, tunelConnet)
}

func (l *Listener) Close() {
	l.CloseOnce.Do(func() {
		close(l.close)
		if l.inner != nil {
			l.inner.Close()
		}
	})
}
