package util

import (
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

// Creates a grpc connection using this projects gRPC version, avoiding dependency issues
func NewGrpcConn(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, grpc.WithInsecure())
}

// Generates a unique id
func Uid() string {
	return uuid.NewV4().String()
}
