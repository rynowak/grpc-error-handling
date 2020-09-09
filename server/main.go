/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a simple gRPC server that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It implements the route guide service whose definition can be found in routeguide/route_guide.proto.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/examples/data"
	"google.golang.org/grpc/status"

	greet "github.com/rynowak/grpc-error-handling/server/hello"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 5000, "The server port")
)

type server struct {
}

func (s *server) SayHello(ctx context.Context, req *greet.HelloRequest) (*greet.HelloReply, error) {

	matched, err := regexp.Match("^[a-zA-Z0-9]*$", []byte(req.Name))
	if err != nil {
		return nil, err
	}

	if matched {
		return &greet.HelloReply{Message: fmt.Sprintf("Hey, %s", req.Name)}, nil
	}

	desc := "The username must only contain alphanumeric characters"
	v := &errdetails.BadRequest_FieldViolation{
		Field:       "username",
		Description: desc,
	}
	br := &errdetails.BadRequest{}
	br.FieldViolations = append(br.FieldViolations, v)
	st := status.New(codes.InvalidArgument, "bad data, try again")
	st, err = st.WithDetails(br)
	if err != nil {
		return nil, err
	}

	return nil, st.Err()
}

func newServer() *server {
	s := &server{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = data.Path("x509/server_cert.pem")
		}
		if *keyFile == "" {
			*keyFile = data.Path("x509/server_key.pem")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	greet.RegisterGreeterServer(grpcServer, newServer())

	fmt.Printf("now listening on port: %d\n", *port)
	grpcServer.Serve(lis)
}
