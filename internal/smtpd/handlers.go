package smtpd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
)

// commandHandler represents a handler for an SMTP command
type commandHandler func(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) (continueLoop bool)

// commandHandlers maps SMTP commands to their handler functions
var commandHandlers = map[string]commandHandler{
	"HELO":     handleHELO,
	"EHLO":     handleEHLO,
	"MAIL":     handleMAIL,
	"RCPT":     handleRCPT,
	"DATA":     handleDATA,
	"QUIT":     handleQUIT,
	"RSET":     handleRSET,
	"NOOP":     handleNOOP,
	"HELP":     handleHELP,
	"VRFY":     handleVRFY,
	"EXPN":     handleEXPN,
	"STARTTLS": handleSTARTTLS,
	"AUTH":     handleAUTH,
	"XCLIENT":  handleXCLIENT,
}

// handleHELO processes the HELO command
func handleHELO(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if len(cmd.fields) < 2 {
		s.writef("501 Syntax error in parameters or arguments")
		return true
	}
	s.clientHostname = cmd.fields[1]
	s.writef("250 %s greets %s", s.srv.Hostname, s.clientHostname)
	
	// RFC 2821 section 4.1.4 specifies that HELO has the same effect as RSET
	*from = ""
	*gotFrom = false
	*to = nil
	*hasRejectedRecipients = false
	buffer.Reset()
	return true
}

// handleEHLO processes the EHLO command
func handleEHLO(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if len(cmd.fields) < 2 {
		s.writef("501 Syntax error in parameters or arguments")
		return true
	}
	s.clientHostname = cmd.fields[1]
	s.esmtp = true
	
	// Create EHLO response
	response := []string{
		fmt.Sprintf("250-%s greets %s", s.srv.Hostname, s.clientHostname),
	}
	
	// Add capabilities
	response = append(response, fmt.Sprintf("250-SIZE %d", s.srv.MaxSize))
	response = append(response, "250-PIPELINING")
	response = append(response, "250-8BITMIME")
	response = append(response, "250-ENHANCEDSTATUSCODES")
	
	if s.srv.TLSConfig != nil && !s.tls {
		response = append(response, "250-STARTTLS")
	}
	
	if s.srv.AuthHandler != nil {
		if s.tls || !s.srv.AuthRequired {
			mechs := s.authMechs()
			if len(mechs) > 0 {
				response = append(response, fmt.Sprintf("250-AUTH %s", strings.Join(mechs, " ")))
			}
		}
	}
	
	// Add final response
	response[len(response)-1] = strings.Replace(response[len(response)-1], "250-", "250 ", 1)
	
	for _, line := range response {
		s.writef("%s", line)
	}
	
	// RFC 2821 section 4.1.4 specifies that EHLO has the same effect as RSET
	*from = ""
	*gotFrom = false
	*to = nil
	*hasRejectedRecipients = false
	buffer.Reset()
	return true
}

// handleMAIL processes the MAIL FROM command
func handleMAIL(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if !s.srv.TLSRequired || s.tls {
		if s.srv.AuthRequired && s.srv.AuthHandler != nil && !s.authenticated {
			s.writef("530 5.7.1 Authentication required")
			return true
		}
	}
	
	if len(cmd.fields) < 2 {
		s.writef("501 5.5.4 Syntax error in parameters or arguments")
		return true
	}
	
	if *gotFrom {
		s.writef("503 5.5.1 Sender already specified")
		return true
	}
	
	mailFrom, params, err := parseMailFrom(strings.Join(cmd.fields[1:], " "))
	if err != nil {
		s.writef("501 5.5.4 Syntax error in parameters or arguments")
		return true
	}
	
	// Check SIZE parameter if present
	if size, ok := params["SIZE"]; ok && s.srv.MaxSize > 0 {
		if parseSize(size) > s.srv.MaxSize {
			s.writef("552 5.3.4 Message size exceeds fixed maximum message size")
			return true
		}
	}
	
	*from = mailFrom
	*gotFrom = true
	s.writef("250 2.1.0 Ok")
	return true
}

// handleRCPT processes the RCPT TO command
func handleRCPT(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if !s.srv.TLSRequired || s.tls {
		if s.srv.AuthRequired && s.srv.AuthHandler != nil && !s.authenticated {
			s.writef("530 5.7.1 Authentication required")
			return true
		}
	}
	
	if !*gotFrom {
		s.writef("503 5.5.1 Bad sequence of commands")
		return true
	}
	
	if len(cmd.fields) < 2 {
		s.writef("501 5.5.4 Syntax error in parameters or arguments")
		return true
	}
	
	rcptTo, _, err := parseRcptTo(strings.Join(cmd.fields[1:], " "))
	if err != nil {
		s.writef("501 5.5.4 Syntax error in parameters or arguments")
		return true
	}
	
	// Check recipient limit
	if len(*to) >= s.srv.MaxRecipients {
		s.writef("452 4.5.3 Too many recipients")
		return true
	}
	
	// Check with recipient handler if configured
	if s.srv.RcptHandler != nil {
		if !s.srv.RcptHandler(s.remoteAddr, *from, rcptTo) {
			if s.srv.IgnoreRejectedRecipients {
				*hasRejectedRecipients = true
				s.writef("250 2.1.5 Ok")
			} else {
				s.writef("550 5.7.1 Recipient rejected")
			}
			return true
		}
	}
	
	*to = append(*to, rcptTo)
	s.writef("250 2.1.5 Ok")
	return true
}

// handleDATA processes the DATA command
func handleDATA(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if !*gotFrom || (len(*to) == 0 && (!s.srv.IgnoreRejectedRecipients || !*hasRejectedRecipients)) {
		s.writef("503 5.5.1 Bad sequence of commands")
		return true
	}
	
	s.writef("354 End data with <CR><LF>.<CR><LF>")
	
	// Read message data
	data, err := s.readData()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			s.writef("421 4.4.2 %s %s ESMTP Service closing transmission channel after timeout exceeded", 
				s.srv.Hostname, s.srv.AppName)
		} else {
			s.writef("451 4.3.0 Requested action aborted: local error in processing")
		}
		return false
	}
	
	// Process the message
	if s.srv.Handler != nil {
		msgID := ""
		if s.srv.HandlerRcpt != nil {
			msgID = s.srv.HandlerRcpt(s.remoteAddr, *from, *to, data)
		} else {
			msgID = s.srv.Handler(s.remoteAddr, *from, *to, data)
		}
		
		if msgID != "" {
			if s.srv.MsgIDHandler != nil {
				s.srv.MsgIDHandler(msgID)
			}
			s.writef("250 2.0.0 Ok: queued as %s", msgID)
		} else {
			s.writef("554 5.7.1 Error: transaction failed")
		}
	} else {
		s.writef("502 5.5.2 Error: command not implemented")
	}
	
	// Reset for next message
	*from = ""
	*gotFrom = false
	*to = nil
	*hasRejectedRecipients = false
	buffer.Reset()
	return true
}

// handleQUIT processes the QUIT command
func handleQUIT(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	s.writef("221 2.0.0 %s %s ESMTP Service closing transmission channel", s.srv.Hostname, s.srv.AppName)
	return false
}

// handleRSET processes the RSET command
func handleRSET(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	*from = ""
	*gotFrom = false
	*to = nil
	*hasRejectedRecipients = false
	buffer.Reset()
	s.writef("250 2.0.0 Ok")
	return true
}

// handleNOOP processes the NOOP command
func handleNOOP(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	s.writef("250 2.0.0 Ok")
	return true
}

// handleHELP processes the HELP command
func handleHELP(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	s.writef("214 2.0.0 This server supports the following commands:")
	s.writef("214 2.0.0 HELO EHLO MAIL RCPT DATA RSET NOOP QUIT HELP VRFY EXPN")
	if s.srv.TLSConfig != nil && !s.tls {
		s.writef("214 2.0.0 STARTTLS")
	}
	if s.srv.AuthHandler != nil {
		s.writef("214 2.0.0 AUTH")
	}
	return true
}

// handleVRFY processes the VRFY command
func handleVRFY(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	s.writef("252 2.1.5 Cannot VRFY user, but will accept message and attempt delivery")
	return true
}

// handleEXPN processes the EXPN command
func handleEXPN(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	s.writef("252 2.1.5 Cannot EXPN list, but will accept message and attempt delivery")
	return true
}

// handleSTARTTLS processes the STARTTLS command
func handleSTARTTLS(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if s.tls {
		s.writef("503 5.5.1 Bad sequence of commands")
		return true
	}
	
	if s.srv.TLSConfig == nil {
		s.writef("502 5.5.2 Error: command not implemented")
		return true
	}
	
	s.writef("220 2.0.0 Ready to start TLS")
	
	// Upgrade connection to TLS
	tlsConn := tls.Server(s.conn, s.srv.TLSConfig)
	if err := tlsConn.Handshake(); err != nil {
		s.writef("403 4.7.0 TLS handshake failed")
		return false
	}
	
	// Replace connection
	s.conn = tlsConn
	s.br = s.br.Reset(s.conn)
	s.bw = s.bw.Reset(s.conn)
	s.tls = true
	
	// Reset session state after STARTTLS
	s.esmtp = false
	s.clientHostname = ""
	return true
}

// handleAUTH processes the AUTH command
func handleAUTH(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if s.srv.AuthHandler == nil {
		s.writef("502 5.5.2 Error: command not implemented")
		return true
	}
	
	if s.authenticated {
		s.writef("503 5.5.1 Already authenticated")
		return true
	}
	
	if !s.tls && s.srv.AuthRequired {
		s.writef("530 5.7.0 Must issue a STARTTLS command first")
		return true
	}
	
	// Process authentication based on mechanism
	if len(cmd.fields) < 2 {
		s.writef("501 5.5.4 Syntax error in parameters or arguments")
		return true
	}
	
	mechanism := strings.ToUpper(cmd.fields[1])
	
	// Handle different auth mechanisms
	switch mechanism {
	case "PLAIN":
		return s.handleAuthPlain(cmd)
	case "LOGIN":
		return s.handleAuthLogin(cmd)
	case "CRAM-MD5":
		return s.handleAuthCramMD5(cmd)
	default:
		s.writef("504 5.5.4 Unrecognized authentication type")
	}
	
	return true
}

// handleXCLIENT processes the XCLIENT command
func handleXCLIENT(s *session, cmd command, from *string, gotFrom *bool, to *[]string, hasRejectedRecipients *bool, buffer *bytes.Buffer) bool {
	if !s.srv.EnableXCLIENT {
		s.writef("502 5.5.2 Error: command not implemented")
		return true
	}
	
	// Check if client is allowed to use XCLIENT
	allowed := false
	for _, ip := range s.srv.XClientAllowed {
		if s.remoteAddr.String() == ip {
			allowed = true
			break
		}
	}
	
	if !allowed {
		s.writef("550 5.7.1 XCLIENT not allowed from your IP")
		return true
	}
	
	// Parse XCLIENT parameters
	for i := 1; i < len(cmd.fields); i++ {
		parts := strings.SplitN(cmd.fields[i], "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		switch strings.ToUpper(parts[0]) {
		case "ADDR":
			// Update remote address
			if parts[1] != "[UNAVAILABLE]" {
				if addr, err := net.ResolveTCPAddr("tcp", parts[1]+":0"); err == nil {
					s.remoteAddr = addr
				}
			}
		case "NAME":
			// Update hostname if available
			if parts[1] != "[UNAVAILABLE]" {
				s.clientHostname = parts[1]
			}
		}
	}
	
	s.writef("250 2.0.0 Ok")
	return true
}