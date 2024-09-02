package main

import "hip/common"


func main() {
	sessionMgr := NewSessionManager()
	listener := NewListener(&common.ProxyProtocol{
		ClientId:         "test-client",
		PublicProtocol:   "tcp",
		PublicIP:         "0.0.0.0",
		//暴露在公网的端口
		PublicPort:       30000,
		InternalProtocol: "tcp",
		//转发到内网的端口
		IternalIp:        "127.0.0.1",
		InternalPort:     3000,
	}, sessionMgr)
	defer listener.Close()
	go func() {
		err := listener.ListenAndServer()
		if err != nil {
			panic(err)
		}
	}()
	//提供内网机器连接公网机器的服务端
	server := NewServer(":35000", sessionMgr)
	err := server.ListenAndServer()
	if err != nil {
		panic(err)
	}
}
