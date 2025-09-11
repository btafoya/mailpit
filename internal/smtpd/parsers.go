package smtpd

import (
	"fmt"
	"strconv"
	"strings"
)

// parseMailFrom parses a MAIL FROM command and returns the sender address and parameters
func parseMailFrom(input string) (string, map[string]string, error) {
	params := make(map[string]string)
	
	// Remove "FROM:" prefix if present
	input = strings.TrimSpace(input)
	if strings.HasPrefix(strings.ToUpper(input), "FROM:") {
		input = strings.TrimSpace(input[5:])
	}
	
	// Find the email address (between < and >)
	startIdx := strings.Index(input, "<")
	endIdx := strings.Index(input, ">")
	
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return "", params, fmt.Errorf("invalid MAIL FROM syntax")
	}
	
	mailFrom := input[startIdx+1 : endIdx]
	
	// Parse parameters after the email address
	if endIdx+1 < len(input) {
		paramStr := strings.TrimSpace(input[endIdx+1:])
		paramPairs := strings.Fields(paramStr)
		
		for _, pair := range paramPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				params[strings.ToUpper(parts[0])] = parts[1]
			}
		}
	}
	
	return mailFrom, params, nil
}

// parseRcptTo parses a RCPT TO command and returns the recipient address and parameters
func parseRcptTo(input string) (string, map[string]string, error) {
	params := make(map[string]string)
	
	// Remove "TO:" prefix if present
	input = strings.TrimSpace(input)
	if strings.HasPrefix(strings.ToUpper(input), "TO:") {
		input = strings.TrimSpace(input[3:])
	}
	
	// Find the email address (between < and >)
	startIdx := strings.Index(input, "<")
	endIdx := strings.Index(input, ">")
	
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return "", params, fmt.Errorf("invalid RCPT TO syntax")
	}
	
	rcptTo := input[startIdx+1 : endIdx]
	
	// Parse parameters after the email address
	if endIdx+1 < len(input) {
		paramStr := strings.TrimSpace(input[endIdx+1:])
		paramPairs := strings.Fields(paramStr)
		
		for _, pair := range paramPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				params[strings.ToUpper(parts[0])] = parts[1]
			}
		}
	}
	
	return rcptTo, params, nil
}

// parseSize parses a SIZE parameter value and returns the size in bytes
func parseSize(sizeStr string) int64 {
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0
	}
	return size
}