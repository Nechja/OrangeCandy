package dap

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	godap "github.com/google/go-dap"
)

type Transport struct {
	reader *bufio.Reader
	writer io.Writer
	logger *slog.Logger
	seq    atomic.Int64
	mu     sync.Mutex
}

func NewTransport(r io.Reader, w io.Writer, logger *slog.Logger) *Transport {
	return &Transport{
		reader: bufio.NewReader(r),
		writer: w,
		logger: logger,
	}
}

func (t *Transport) NextSeq() int {
	return int(t.seq.Add(1))
}

func (t *Transport) Send(message godap.Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Debug("dap send", "seq", message.GetSeq(), "type", fmt.Sprintf("%T", message))
	return godap.WriteProtocolMessage(t.writer, message)
}

func (t *Transport) Receive() (godap.Message, error) {
	msg, err := godap.ReadProtocolMessage(t.reader)
	if err != nil {
		return nil, err
	}
	t.logger.Debug("dap recv", "seq", msg.GetSeq(), "type", fmt.Sprintf("%T", msg))
	return msg, nil
}
