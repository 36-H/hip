package main

import (
	"flag"
	"hip/common"
)

func main() {
	var confFile string
	flag.StringVar(&confFile, "c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(confFile)
	if err != nil {
		panic(err)
	}

	listenerConfigs, err := ParseListenerConfig(conf.ListenerFile)
	if err != nil {
		panic(err)
	}

	sessionMgr := NewSessionManager()

	for _, listenerConfig := range listenerConfigs {
		listener := NewListener(&common.ProxyProtocol{
			ClientId:         listenerConfig.ClientID,
			PublicProtocol:   listenerConfig.PublicProtocol,
			PublicIP:         listenerConfig.PublicIP,
			PublicPort:       listenerConfig.PublicPort,
			InternalProtocol: listenerConfig.InternalProtocol,
			InternalIp:       listenerConfig.InternalIP,
			InternalPort:     listenerConfig.InternalPort,
		}, sessionMgr)
		go func() {
			defer listener.Close()
			err := listener.ListenAndServer()
			if err != nil {
				panic(err)
			}
		}()
	}
	//提供内网机器连接公网机器的服务端
	server := NewServer(":35000", sessionMgr)
	err = server.ListenAndServer()
	if err != nil {
		panic(err)
	}
}
