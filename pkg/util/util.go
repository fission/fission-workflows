package util

import (
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"fmt"
)

// Creates a grpc connection using this projects gRPC version, avoiding dependency issues
func NewGrpcConn(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, grpc.WithInsecure())
}

// Generates a unique id
func Uid() string {
	return uuid.NewV4().String()
}


func CreateScopeId(parentId string, taskId string) string {
	if len(taskId) == 0 {
		taskId = Uid()
	}
	return fmt.Sprintf("%s_%s", parentId, taskId)
}
