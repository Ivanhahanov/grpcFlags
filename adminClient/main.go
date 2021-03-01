package main

import (
	"context"
	"fmt"
	pb "github.com/Ivanhahanov/grpcFlags/flags"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

const address = "localhost:50051"

func GetListOfUsers(conn *grpc.ClientConn) {
	c := pb.NewFlagsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ListOfUsers(ctx, &pb.Admin{Username: "admin", Password: "adminSuperSecretPassword"})
	if err != nil {
		log.Fatalf("could not send flag: %v", err)
	}
	fmt.Printf("User\tTeam\tChallenges\n")
	var services []string
	for {
		cur, err := r.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			continue
		}
		for _, service := range cur.Services {
			if service.Status {
				services = append(services, service.Name)
			}
		}
		fmt.Printf("%s\t%s\t%v\n", cur.Username, cur.Team, services)
	}
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	GetListOfUsers(conn)
}
