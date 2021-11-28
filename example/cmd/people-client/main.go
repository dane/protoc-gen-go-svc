package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	v1publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	v2publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

func main() {
	var (
		isV2 bool
		addr string
	)

	flag.BoolVar(&isV2, "v2", false, "use v2 client")
	flag.StringVar(&addr, "addr", ":8000", "service address")
	flag.Parse()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	clientV1 := v1publicpb.NewPeopleClient(conn)
	clientV2 := v2publicpb.NewPeopleClient(conn)

	if isV2 {
		res, err := clientV2.Create(ctx, &v2publicpb.CreateRequest{
			Id:         uuid.New().String(),
			FullName:   "Dane Harrigan",
			Employment: v2publicpb.Person_FULL_TIME,
			Age:        36,
			Hobby: &v2publicpb.Hobby{
				Type: &v2publicpb.Hobby_Cycling{
					Cycling: &v2publicpb.Cycling{
						Style: "road",
					},
				},
			},
		})

		if err != nil {
			fmt.Printf("V2 client error: %q\n", err)
			return
		}

		fmt.Printf("V1 res: %v", res)
		return
	}

	res, err := clientV1.Create(ctx, &v1publicpb.CreateRequest{
		Id:         uuid.New().String(),
		FirstName:  "Dane",
		LastName:   "Harrigan",
		Employment: v1publicpb.Person_EMPLOYED,
		Hobby: &v1publicpb.Hobby{
			Type: &v1publicpb.Hobby_Biking{
				Biking: &v1publicpb.Biking{
					Style: "road",
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("V1 client error: %q\n", err)
		return
	}

	fmt.Printf("V1 res: %v", res)
}
