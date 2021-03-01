package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	pb "github.com/Ivanhahanov/grpcFlags/flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"time"
)

const address = "localhost:50051"

var secretKey string
var tempFlag string

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func SubmitFlag(conn *grpc.ClientConn) {
	c := pb.NewFlagsClient(conn)
	//flag := tempFlag
	flag := "Flag{h9h2fhfUVuS9jZ8uVbhV}"
	encryptedFlag, _ := encrypt([]byte(flag), []byte(secretKey))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, "username", "user")
	defer cancel()
	r, err := c.CheckFlag(ctx, &pb.CheckFlagRequest{
		Flag:    encryptedFlag,
		Service: "sql",
	})
	if err != nil {
		log.Fatalf("could not send flag: %v", err)
	}
	log.Printf("Check Flag: %t", r.Status)
}

func AddService(conn *grpc.ClientConn) {
	c := pb.NewFlagsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, "username", "user")
	defer cancel()
	r, err := c.ServiceRegister(ctx, &pb.ServiceRegisterRequest{
		Service: "admin",
		Key:     "adminSuperSecretKey",
	})
	if err != nil {
		log.Fatalf("could not send request: %v", err)
	}
	decryptedFlag, _ := decrypt(r.Flag, []byte(secretKey))
	tempFlag = string(decryptedFlag)
	log.Printf("Flag: %s", tempFlag)
}

func RegisterUser(conn *grpc.ClientConn) {
	c := pb.NewFlagsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.UserRegister(ctx, &pb.UserRegisterRequest{
		Username: "user",
		Password: "userSuperSecretKey",
		Team:     "Test",
	})
	if err != nil {
		log.Fatalf("sorry, %v", err)
	}
	log.Printf("Hello %s, from %s!", r.Username, r.Team)
}

func GetKey(conn *grpc.ClientConn) {
	c := pb.NewFlagsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetKey(ctx, &pb.GetKeyRequest{
		Username: "user",
		Password: "userSuperSecretKey",
	})
	if err != nil {
		log.Fatalf("sorry, %v", err)
	}
	secretKey = r.Key
	log.Printf("Secret key loading successfuly")
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//RegisterUser(conn)
	GetKey(conn)
	//AddService(conn)
	SubmitFlag(conn)
	conn.Close()
}
