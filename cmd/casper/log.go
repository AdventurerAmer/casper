package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type LogOp uint8

const (
	LogOpPut LogOp = iota
	LogOpDelete
)

type LogEvent struct {
	Op    LogOp
	Key   string
	Value []byte
}

type LogEntry struct {
	SequanceNum uint64
	Op          LogOp
	Key         string
	Value       []byte
}

func (op LogOp) String() string {
	switch op {
	case LogOpPut:
		return "PUT"
	case LogOpDelete:
		return "DELETE"
	}
	return ""
}

func (e LogEntry) String() string {
	var v any
	err := json.Unmarshal(e.Value, &v)
	if err != nil {
		panic(err)
	}
	if e.Op == LogOpPut {
		return fmt.Sprintf("%d. [%s] (%s, %v)", e.SequanceNum, e.Op, e.Key, v)
	}
	return fmt.Sprintf("%d. [%s] (%s)", e.SequanceNum, e.Op, e.Key)
}

type WALogger struct {
	File        *os.File
	LogEventCh  chan LogEvent
	SequanceNum uint64
}

func NewWALogger() (*WALogger, error) {
	filename := "wal000.log"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644) // 0644 read/write for owner ready only for others
	if err != nil {
		return nil, fmt.Errorf("failed to create or open log file '%s': %w", filename, err)
	}
	logger := &WALogger{
		File:       file,
		LogEventCh: make(chan LogEvent, 64),
	}
	return logger, nil
}

func (l *WALogger) Close() error {
	close(l.LogEventCh)
	err := l.File.Close()
	return err
}

func (l *WALogger) Put(k string, v []byte) {
	l.LogEventCh <- LogEvent{Op: LogOpPut, Key: k, Value: v}
}

func (l *WALogger) Delete(k string) {
	l.LogEventCh <- LogEvent{Op: LogOpDelete, Key: k}
}

func ReadEntires(file *os.File) []LogEntry {
	var entires []LogEntry
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := uint64(info.Size())
	buf := make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	offset := uint64(0)
	for offset < size {
		seqNum := binary.LittleEndian.Uint64(buf[offset : offset+8])
		offset += 8
		op := LogOp(buf[offset])
		offset += 1
		keyLength := binary.LittleEndian.Uint64(buf[offset : offset+8])
		offset += 8
		key := buf[offset : offset+keyLength]
		offset += keyLength
		var value []byte
		if op == LogOpPut {
			valueLength := binary.LittleEndian.Uint64(buf[offset : offset+8])
			offset += 8

			value = buf[offset : offset+valueLength]
			offset += valueLength
		}
		entires = append(entires, LogEntry{SequanceNum: seqNum, Op: op, Key: string(key), Value: value})
	}
	return entires
}

func (l *WALogger) WriteLoop() {
	var buf bytes.Buffer
	for {
		select {
		case event, ok := <-l.LogEventCh:
			if !ok {
				return
			}
			seqNum := l.SequanceNum
			l.SequanceNum++
			binary.Write(&buf, binary.LittleEndian, seqNum)
			binary.Write(&buf, binary.LittleEndian, event.Op)
			keyLength := uint64(len(event.Key))
			binary.Write(&buf, binary.LittleEndian, keyLength)
			buf.WriteString(event.Key)
			if event.Op == LogOpPut {
				valueLength := uint64(len(event.Value))
				binary.Write(&buf, binary.LittleEndian, valueLength)
				buf.Write(event.Value)
			}
			_, err := l.File.Write(buf.Bytes())
			if err != nil {
				log.Fatal(err)
			}
			err = l.File.Sync()
			if err != nil {
				log.Fatal(err)
			}
			buf.Reset()
		}
	}
}
