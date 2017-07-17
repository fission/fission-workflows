#!/bin/bash

protoc -I . pkg/workflow-engine/workflow-engine.proto --go_out=plugins=grpc:.
