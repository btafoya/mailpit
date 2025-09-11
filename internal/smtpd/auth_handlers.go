package smtpd

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"
)

// handleAuthPlain handles PLAIN authentication
func (s *session) handleAuthPlain(cmd command) bool {
	var auth string
	
	if len(cmd.fields) > 2 {
		// Initial response provided
		auth = strings.Join(cmd.fields[2:], " ")
	} else {
		// Request initial response
		s.writef("334 ")
		line, err := s.readLine()
		if err != nil {
			s.writef("454 4.7.0 Temporary authentication failure")
			return true
		}
		auth = line
	}
	
	if auth == "*" {
		s.writef("501 5.7.0 Authentication cancelled")
		return true
	}
	
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		s.writef("501 5.5.2 Cannot decode base64")
		return true
	}
	
	// Parse PLAIN auth (NULL separated: authzid \x00 authcid \x00 passwd)
	parts := strings.Split(string(data), "\x00")
	if len(parts) != 3 {
		s.writef("501 5.5.2 Invalid authentication data")
		return true
	}
	
	username := parts[1]
	password := parts[2]
	
	if s.srv.AuthHandler(s.remoteAddr, "PLAIN", username, password, "") {
		s.authenticated = true
		s.authUser = username
		s.writef("235 2.7.0 Authentication successful")
	} else {
		s.writef("535 5.7.8 Authentication credentials invalid")
	}
	
	return true
}

// handleAuthLogin handles LOGIN authentication
func (s *session) handleAuthLogin(cmd command) bool {
	var username, password string
	
	// Check if username is provided with command
	if len(cmd.fields) > 2 {
		username = strings.Join(cmd.fields[2:], " ")
	} else {
		// Request username
		s.writef("334 VXNlcm5hbWU6") // "Username:" in base64
		line, err := s.readLine()
		if err != nil {
			s.writef("454 4.7.0 Temporary authentication failure")
			return true
		}
		if line == "*" {
			s.writef("501 5.7.0 Authentication cancelled")
			return true
		}
		username = line
	}
	
	// Decode username
	userBytes, err := base64.StdEncoding.DecodeString(username)
	if err != nil {
		s.writef("501 5.5.2 Cannot decode base64")
		return true
	}
	username = string(userBytes)
	
	// Request password
	s.writef("334 UGFzc3dvcmQ6") // "Password:" in base64
	line, err := s.readLine()
	if err != nil {
		s.writef("454 4.7.0 Temporary authentication failure")
		return true
	}
	if line == "*" {
		s.writef("501 5.7.0 Authentication cancelled")
		return true
	}
	
	// Decode password
	passBytes, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		s.writef("501 5.5.2 Cannot decode base64")
		return true
	}
	password = string(passBytes)
	
	if s.srv.AuthHandler(s.remoteAddr, "LOGIN", username, password, "") {
		s.authenticated = true
		s.authUser = username
		s.writef("235 2.7.0 Authentication successful")
	} else {
		s.writef("535 5.7.8 Authentication credentials invalid")
	}
	
	return true
}

// handleAuthCramMD5 handles CRAM-MD5 authentication
func (s *session) handleAuthCramMD5(cmd command) bool {
	// Generate challenge
	challenge := fmt.Sprintf("<%d.%d@%s>", 
		s.conn.RemoteAddr().(*net.TCPAddr).Port,
		time.Now().Unix(),
		s.srv.Hostname)
	
	// Send challenge
	s.writef("334 %s", base64.StdEncoding.EncodeToString([]byte(challenge)))
	
	// Read response
	line, err := s.readLine()
	if err != nil {
		s.writef("454 4.7.0 Temporary authentication failure")
		return true
	}
	
	if line == "*" {
		s.writef("501 5.7.0 Authentication cancelled")
		return true
	}
	
	// Decode response
	data, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		s.writef("501 5.5.2 Cannot decode base64")
		return true
	}
	
	// Parse response (username and digest separated by space)
	parts := strings.SplitN(string(data), " ", 2)
	if len(parts) != 2 {
		s.writef("501 5.5.2 Invalid authentication data")
		return true
	}
	
	username := parts[0]
	digest := parts[1]
	
	// Verify with handler (handler should compute expected digest)
	if s.srv.AuthHandler(s.remoteAddr, "CRAM-MD5", username, challenge, digest) {
		s.authenticated = true
		s.authUser = username
		s.writef("235 2.7.0 Authentication successful")
	} else {
		s.writef("535 5.7.8 Authentication credentials invalid")
	}
	
	return true
}

// authMechs returns the list of supported authentication mechanisms
func (s *session) authMechs() []string {
	mechs := []string{}
	
	if s.srv.AuthHandler != nil {
		// Always support PLAIN and LOGIN over TLS
		if s.tls {
			mechs = append(mechs, "PLAIN", "LOGIN")
		}
		// CRAM-MD5 can be used without TLS
		mechs = append(mechs, "CRAM-MD5")
	}
	
	return mechs
}

// Helper function to compute CRAM-MD5 digest (for testing)
func computeCramMD5(username, password, challenge string) string {
	h := hmac.New(md5.New, []byte(password))
	h.Write([]byte(challenge))
	digest := fmt.Sprintf("%s %x", username, h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(digest))
}