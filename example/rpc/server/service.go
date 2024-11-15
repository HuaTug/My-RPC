package main

import (
	"context"
	"fmt"
)

func (u *UserService) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		Name: fmt.Sprintf("%d-%s", req.Id, "Hello, I'm user"),
	}, nil
}

func (u *UserParentService) GetParentById(ctx context.Context, req *GetParentByIdReq) (*GetParentByIdResp, error) {
	return &GetParentByIdResp{
		Father: fmt.Sprintf("%d-%s", req.Id, "Hello, I'm father"),
		Mother: fmt.Sprintf("%d-%s", req.Id, "Hello, I'm mother"),
	}, nil
}
