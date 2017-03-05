package main

import (
	"net"
)

// UMServer is local tcp service listening for urgent messages
type UMServer struct {
	Addr    string
	Message chan string
}

// Start local tcp service
func (s *UMServer) Start() {
	if s.Addr == "" {
		s.Addr = ":8681"
	}

	logger.Infof("UMServer.Start starting (%s)...", s.Addr)

	s.Message = make(chan string)

	go func() {
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			logger.Error("UMServer.Start Listen error: ", err.Error())
			return
		}
		defer func() {
			if ln != nil {
				ln.Close()
			}
		}()
		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Error("UMServer.Start Error accepting: ", err.Error())
				return
			}
			buf := make([]byte, 1024)
			if reqLen, err := conn.Read(buf); err == nil || reqLen > 0 {
				s.Message <- string(buf[:reqLen])
			}
			conn.Close()
		}
	}()
}
