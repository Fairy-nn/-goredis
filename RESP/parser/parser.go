package parser

import (
	"bufio"
	"errors"
	"fmt"
	resp "goredis/RESP/interface"
	"io"
	"strconv"
)

// store the parser result
type Parser struct {
	Data resp.Reply
	Err  error
}

type ParserResult struct {
	readingMultiLine  bool     // whether reading multi line
	expectedArgsCount int      // expected args count
	msgType           byte     // message type
	args              [][]byte // args
	bulkLen           int64    // bulk length
}

// func (p *ParserResult) isDone() bool {
// 	return p.expectedArgsCount > 0 && len(p.args) == p.expectedArgsCount
// }

// parses the RESP stream and returns a channel of Parser results
func ParseStream(reader io.Reader) <-chan *Parser {
	ch := make(chan *Parser, 1)
	go parseIt(reader, ch)
	return ch
}

// readLine reads a line from the bufio.Reader and returns the line, a boolean indicating if it's an io error, and an error if any
func ReadLine(bufReader *bufio.Reader, state *ParserResult) ([]byte, bool, error) {
	var line []byte
	var err error
	// read the first byte to determine the message type
	if state.bulkLen == 0 {
		line, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err // io.EOF
		}
		if len(line) == 0 || line[len(line)-2] != '\r' {
			return nil, false, errors.New("invalid line terminator")
		}
	} else {
		// read the bulk length
		line = make([]byte, state.bulkLen+2) // +2 for \r\n
		_, err = io.ReadFull(bufReader, line)
		if err != nil {
			return nil, true, err // io.EOF
		}
		if line[len(line)-2] != '\r' || line[len(line)-1] != '\n' || len(line) != int(state.bulkLen+2) {
			return nil, false, errors.New("invalid bulk terminator")
		}
		state.bulkLen = 0 // reset bulk length
	}
	return line, false, nil
}

// parse mutiline header
func parseMultiBulkHeader(msg []byte, state *ParserResult) error {
	var err error
	var expectedArgsCount uint64
	expectedArgsCount, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("Protocol error" + string(msg))
	}
	if expectedArgsCount == 0 { // empty multi-bulk reply
		state.expectedArgsCount = 0
		return nil
	} else if expectedArgsCount > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedArgsCount)
		state.args = make([][]byte, expectedArgsCount)
		return nil
	}
	return errors.New("Protocol error" + string(msg))
}

// parse single line header
func parseBulkHeader(msg []byte, state *ParserResult) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("Protocol error" + string(msg))
	}
	if state.bulkLen == -1 { // null bulk reply
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	}
	return errors.New("Protocol error" + string(msg))
}

// uses a goroutine to parse the RESP stream and send results to the channel
func parseIt(reader io.Reader, ch chan<- *Parser) {
	// panic recovery to handle unexpected errors
	defer func() {
		if r := recover(); r != nil {
			ch <- &Parser{Err: io.EOF}
		}
	}()
	//
	bufReader := bufio.NewReader(reader)
	var parserResult ParserResult
	var err error
	var msg []byte
	for {
		var ioErr bool
		msg, ioErr, err = ReadLine(bufReader, &parserResult)
		if err != nil {
			// io error, return the error
			if ioErr {
				ch <- &Parser{Err: err}
				close(ch)
				return
			}
			ch <- &Parser{Err: err}
			parserResult = ParserResult{}
			continue
		}
		// multiline message
		if !parserResult.readingMultiLine {
			if msg[0] == '*' { // represents a multi-bulk reply
				// parse the number of arguments
				err = parseMultiBulkHeader(msg, &parserResult)
				if err != nil {
					ch <- &Parser{Err: errors.New("Protocol error" + string(msg))}
					parserResult = ParserResult{}
					continue
				}
				if parserResult.expectedArgsCount == 0 {
					// empty multi-bulk reply
					ch <- &Parser{Data: resp.MakeEmptyMultiBulkReply()}
					parserResult = ParserResult{}
					continue
				}
			}
		} else if msg[0] == '$' { // mutiline message
			err = parseBulkHeader(msg, &parserResult)
			if err != nil {
				ch <- &Parser{Err: errors.New("Protocol error" + string(msg))}
				parserResult = ParserResult{}
				continue
			}
			if parserResult.bulkLen == -1 {
				// null bulk reply
				ch <- &Parser{Data: resp.MakeNullReply()}
				parserResult = ParserResult{}
				continue
			}

		} else { // single line message
			if msg[0] == '+' { // simple string reply
				ch <- &Parser{Data: resp.MakeStatusReply(string(msg[1:]))}
				parserResult = ParserResult{}
				continue
			} else if msg[0] == '-' { // error reply
				ch <- &Parser{Data: resp.MakeStandardErrorReply(string(msg[1:]))}
				parserResult = ParserResult{}
				continue
			} else if msg[0] == ':' { // integer reply
				var code int64
				_, err = fmt.Sscanf(string(msg[1:]), "%d", &code)
				if err != nil {
					ch <- &Parser{Err: errors.New("Protocol error" + string(msg))}
					parserResult = ParserResult{}
					continue
				}
				ch <- &Parser{Data: resp.MakeIntegerReply(code)}
				parserResult = ParserResult{}
				continue
			}

		}

	}
}
