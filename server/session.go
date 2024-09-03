package main

import (
	"fmt"
	"github.com/xtaci/smux"
	"net"
	"sync"
)

type Session struct {
	ClientId   string
	Connection *smux.Session
}

type SessionManager struct {
	lock     sync.RWMutex
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (mgr *SessionManager) GetSesssionByClientId(clientId string) (net.Conn, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	session := mgr.sessions[clientId]
	if session == nil {
		return nil, fmt.Errorf("client [%s] is not connected", clientId)
	}

	stream, err := session.Connection.OpenStream()
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (mgr *SessionManager) CreateSession(clientId string, conn net.Conn) (*Session, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	old := mgr.sessions[clientId]
	if old != nil {
		return nil, fmt.Errorf("client [%s] is online", clientId)
	}

	mux, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ClientId:   clientId,
		Connection: mux,
	}
	mgr.sessions[clientId] = session
	return session, nil
}

func (mgr *SessionManager) CloseSession(clientId string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	session := mgr.sessions[clientId]
	if session == nil {
		return
	}

	session.Connection.Close()
	delete(mgr.sessions, clientId)
}

func (mgr *SessionManager) Range(f func(k string, v *Session) bool) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	for k, v := range mgr.sessions {
		ok := f(k, v)
		if !ok {
			delete(mgr.sessions, k)
		}
	}
}
