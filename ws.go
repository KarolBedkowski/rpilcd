package main

import (
	"log"
	"net"
)

// WSServer is local tcp service listening for urgent messages
type WSServer struct {
	Addr    string
	Message chan string
}

// Start local tcp service
func (s *WSServer) Start() {
	if s.Addr == "" {
		s.Addr = ":8681"
	}

	log.Printf("WSServer.Start starting (%s)...", s.Addr)

	s.Message = make(chan string, 10)

	go func() {
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			log.Printf("WSServer.Start Listen error: %s", err.Error())
			return
		}
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("WSServer.Start Error accepting: %s", err.Error())
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
