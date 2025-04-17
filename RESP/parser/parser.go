package parser

import (
	"bufio"
	"errors"
	"fmt"
	"goredis/interface/resp"
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

func (p *ParserResult) isDone() bool {
	return p.expectedArgsCount > 0 && len(p.args) == p.expectedArgsCount
}

// parses the RESP stream and returns a channel of Parser results
func ParseStream(reader io.Reader) <-chan *Parser {
	ch := make(chan *Parser, 1)
	go parseIt(reader, ch)
	return ch
}

// readLine reads a line from the bufio.Reader and returns the line, a boolean indicating if it's an io error, and an error if any
func readLine(bufReader *bufio.Reader, state *ParserResult) ([]byte, bool, error) {
	var line []byte
	var err error
	if state.bulkLen == 0 {
		line, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err // io.EOF
		}
		if len(line) == 0 || line[len(line)-2] != '\r' {
			return nil, false, errors.New("invalid line terminator")
		}
	} else {
		// 读取批量字符串内容
		line = make([]byte, state.bulkLen+2) // +2 for \r\n
		_, err = io.ReadFull(bufReader, line)
		if err != nil {
			return nil, true, err // io.EOF
		}
		if line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			return nil, false, errors.New("invalid bulk terminator")
		}
		state.bulkLen = 0 // 重置 bulkLen
	}
	return line, false, nil
}

// parse mutiline header
func parseMultiBulkHeader(msg []byte, state *ParserResult) error {
	var err error
	var expectedArgsCount uint64
	expectedArgsCount, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
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
		state.args = make([][]byte, 0, expectedArgsCount) //一开始这里用的是 make([][]byte, expectedArgsCount) 但是这样会导致在读取多行数据时，args的长度不够，导致数组越界
		// 这里改成了 make([][]byte, 0, expectedArgsCount) 这样就不会越界了
		return nil
	} else {
		// invalid multi-bulk reply
		return errors.New("Protocol error" + string(msg))
	}
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
	} else {
		return errors.New("Protocol error" + string(msg))
	}

}

//调用 readLine 从输入流中读取一行数据。
// 根据 msg[0] 判断消息类型：
// *：多条批量字符串，调用 parseMultiBulkHeader。
// $：批量字符串，调用 parseBulkHeader。
// +：简单字符串，直接生成 StatusReply。
// -：错误消息，直接生成 StandardErrorReply。
// :：整数，直接生成 IntegerReply。
// 如果是多行消息，调用 readBody 读取后续行。
// 当 parserResult.isDone() 返回 true 时，表示消息解析完成，将结果发送到 ch

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
		msg, ioErr, err = readLine(bufReader, &parserResult) // read a line from the buffer
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
		// not multiline message
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
		} else { // multiline message
			err = readBody(msg, &parserResult)
			if err != nil {
				ch <- &Parser{
					Err: errors.New("protocol error: " + string(msg)),
				}
				parserResult = ParserResult{} // reset parser result
				continue
			}
			if parserResult.isDone() {
				var result resp.Reply
				if parserResult.msgType == '*' {
					// multi-bulk reply
					result = resp.MakeMultiBulkReply(parserResult.args)
				}
				if parserResult.msgType == '$' {
					// bulk reply
					if parserResult.bulkLen == 0 {
						result = resp.MakeEmptyBulkReply()
					} else {
						result = resp.MakeBulkReply(parserResult.args[0])
					}
				}
				ch <- &Parser{
					Data: result,
					Err:  err,
				}
				parserResult = ParserResult{} // reset parser result
			}
		}
	}
}
func readBody(msg []byte, state *ParserResult) error {
	if len(msg) < 2 {
		return errors.New("protocol error: message too short")
	}
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
