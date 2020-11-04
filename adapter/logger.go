package adapter

import (
	"context"

	"github.com/layer5io/meshkit/logger"
)

type adapterLogger struct {
	log  logger.Handler
	next Handler
}

func AddLogger(logger logger.Handler, h Handler) Handler {
	return &adapterLogger{
		log:  logger,
		next: h,
	}
}

func (s *adapterLogger) GetName() string {
	if !(len(s.next.GetName()) > 1) {
		s.log.Error(ErrGetName)
	}
	return s.next.GetName()
}

func (s *adapterLogger) CreateInstance(b []byte, st string, c *chan interface{}) error {
	s.log.Info("Creating instance")
	err := s.next.CreateInstance(b, st, c)
	if err != nil {
		s.log.Error(err)
	}
	return err
}

func (s *adapterLogger) ApplyOperation(ctx context.Context, op OperationRequest) error {
	s.log.Info("Applying operation ", op.OperationName)
	err := s.next.ApplyOperation(ctx, op)
	if err != nil {
		s.log.Error(err)
	}
	return err
}

func (s *adapterLogger) ListOperations() (Operations, error) {
	s.log.Info("Listing Operations")
	ops, err := s.next.ListOperations()
	if err != nil {
		s.log.Error(err)
	}
	return ops, err
}

func (s *adapterLogger) StreamErr(e *Event, err error) {
	s.log.Error(err)
}

func (s *adapterLogger) StreamInfo(*Event) {
	s.log.Info("Sending event response")
}
