package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

var (
	WriteTimeOut = time.Second * 3
	ReconnetTimeOut = time.Second * 1
)

type ClientInfo struct {
	ClientId         string
	PublicProtocol   string
	PublicIp         string
	PublicPort       uint16
	InternalProtocol string
	InternalIp       string
	InternalPort     uint16
}

var (
	cmdProtocol     = 0x0
	cmdHandshakeReq = 0x1
)

type ProxyProtocol struct {
	ClientId         string
	PublicProtocol   string
	PublicIP         string
	PublicPort       uint16
	InternalProtocol string
	IternalIp        string
	InternalPort     uint16
}

// 1byte version
// 1byte cmd
// 2bytes length
// length bytes body
func (pp *ProxyProtocol) Encode() ([]byte, error) {
	header := make([]byte, 4)
	header[0] = 0x0
	header[1] = byte(cmdProtocol)

	body, err := json.Marshal(pp)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(header[2:4], uint16(len(body)))
	return append(header, body...), nil
}

func (pp *ProxyProtocol) Decode(reader io.Reader) error {
	hdr := make([]byte, 4)
	_, err := io.ReadFull(reader, hdr)
	if err != nil {
		return err
	}
	cmd := hdr[1]
	if cmd != byte(cmdProtocol) {
		return fmt.Errorf("invalid proxyProtocol cmd")
	}

	bodyLen := binary.BigEndian.Uint16(hdr[2:4])
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, pp)
}

type HandshakeReq struct {
	ClientId string
}

func (hr *HandshakeReq) Encode() ([]byte, error) {
	header := make([]byte, 4)
	header[0] = 0x0
	header[1] = byte(cmdHandshakeReq)

	body, err := json.Marshal(hr)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(header[2:4], uint16(len(body)))
	return append(header, body...), nil
}

func (hr *HandshakeReq) Decode(reader io.Reader) error {
	hdr := make([]byte, 4)
	_, err := io.ReadFull(reader, hdr)
	if err != nil {
		return err
	}
	cmd := hdr[1]
	if cmd != byte(cmdHandshakeReq) {
		return fmt.Errorf("invalid handshake cmd")
	}

	bodyLen := binary.BigEndian.Uint16(hdr[2:4])
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, hr)
}
