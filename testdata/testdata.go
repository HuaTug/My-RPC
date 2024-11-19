package testdata

import (
	"context"
	"errors"
)

type Service struct {
}

type HelloRequest struct {
	Msg string
}

type Helloreply struct {
	Msg string
}

func (s *Service) SayHello(ctx context.Context, req *HelloRequest) (*Helloreply, error) {
	rsp := &Helloreply{
		Msg: req.Msg + "world",
	}
	return rsp, nil
}

type CalculatorService struct{}

type CalculateRequest struct {
	Operation string
	Num1      float64
	Num2      float64
}

type CalculateReply struct {
	Result float64
}

func (s *CalculatorService) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateReply, error) {
	var result float64
	switch req.Operation {
	case "add":
		result = req.Num1 + req.Num2
	case "subtract":
		result = req.Num1 - req.Num2
	case "multiply":
		result = req.Num1 * req.Num2
	case "divide":
		if req.Num2 == 0 {
			return nil, errors.New("division by zero")
		}
		result = req.Num1 / req.Num2
	default:
		return nil, errors.New("invalid operation")
	}
	return &CalculateReply{Result: result}, nil
}
