package aof

import (
	"fmt"
	"goredis/config"
	"goredis/interface/database"
	"goredis/lib/logger"
	"goredis/lib/utils"
	"goredis/resp/connection"
	"goredis/resp/parser"
	"goredis/resp/reply"
	"io"
	"os"
	"strconv"
)

const aofBufferSize = 1 << 16

type CmdLine = [][]byte // CmdLine represents a command line in Redis, which is an array of byte slices.

type payload struct {
	cmdLine CmdLine // The command line to be executed.
	dbIndex int     // The index of the database to be used.
}

type AofHandler struct {
	db          database.Database // The database instance to be used for executing commands.
	aofChan     chan *payload
	aofFile     *os.File // The file where the AOF (Append Only File) is stored.
	aofFileName string   // The name of the AOF file.
	currentDB   int
}

func NewAofHandler(db database.Database) (*AofHandler, error) {
	aofhandler := &AofHandler{
		db:          db,
		aofFileName: config.Properties.AppendFilename,
		aofChan:     make(chan *payload, aofBufferSize),
	}
	//	fmt.Println("open aof file: " + aofhandler.aofFileName)
	aofFile, err := os.OpenFile(aofhandler.aofFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logger.Error("open aof file error: " + err.Error())
		return nil, err
	}
	aofhandler.aofFile = aofFile

	aofhandler.LoadAof()

	// goroutine to write aof file
	go func() {
		aofhandler.WriteAofFile()
	}()

	return aofhandler, nil
}

// increase command and dbIndex to aofChan
func (h *AofHandler) AddCommand(dbIndex int, cmdLine CmdLine) {
	if h.aofChan == nil || !config.Properties.AppendOnly {
		h.aofChan = make(chan *payload, 100)
	}

	h.aofChan <- &payload{
		cmdLine: cmdLine,
		dbIndex: dbIndex,
	}
}

// write aof file
func (h *AofHandler) WriteAofFile() error {
	h.currentDB = 0
	for p := range h.aofChan { // read from aofChan
		fmt.Println("select db: " + strconv.Itoa(p.dbIndex))
		fmt.Println("current db: " + strconv.Itoa(h.currentDB))
		if p.dbIndex != h.currentDB {
			h.currentDB = p.dbIndex
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := h.aofFile.Write(data)
			if err != nil {
				logger.Error("write aof file error: " + err.Error())
				continue
			}
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := h.aofFile.Write(data)
		if err != nil {
			logger.Error("write aof file error: " + err.Error())
			continue
		}
	}
	return nil
}

// exec command from aof file
func (h *AofHandler) LoadAof() {
	aofFile, err := os.Open(h.aofFileName)
	if err != nil {
		logger.Error("AOF file open error: " + err.Error())
		return
	}
	defer aofFile.Close()

	ch := parser.ParseStream(aofFile)
	fakeConn := &connection.Connection{}
	for p := range ch {
		if p.Err != nil {
			// If the error is EOF or unexpected EOF, break the loop
			if p.Err == io.EOF || p.Err == io.ErrUnexpectedEOF {
				// End of file
				break
			}
			// Other errors
			logger.Error("AOF file parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("AOF file empty payload")
			continue
		}

		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("AOF file require multi bulk reply")
			continue
		}

		rep := h.db.Exec(fakeConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error("Execute AOF command error")
		}
	}
}
