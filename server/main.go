package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	pb "github.com/Ivanhahanov/grpcFlags/flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
)

const port = ":50051"

type server struct {
	pb.FlagsServer
}
type Users struct {
	Username string
	Team     string
	Password string
	Key      string
	Services []Service
}

type Service struct {
	Name   string
	Flag   string
	Status bool
}

var temporaryUsers = []Users{
	{
		Username: "admin",
		Team:     "admin",
		Password: "adminSuperSecretPassword",
	},
}

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
	if _, err = io.ReadFull(cryptoRand.Reader, nonce); err != nil {
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

func (s *server) CheckFlag(ctx context.Context, in *pb.CheckFlagRequest) (*pb.CheckFlagResult, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	username := md.Get("username")[0]
	userId, user := getUser(username)
	decryptedFlag, _ := decrypt(in.Flag, []byte(user.Key))
	flag := string(decryptedFlag)
	log.Printf("Received: %s", flag)
	if strings.HasPrefix(flag, "Flag{") && strings.HasSuffix(flag, "}") {
		serviceId, service := getService(username, in.Service)
		if service.Flag == flag {
			temporaryUsers[userId].Services[serviceId].Status = true
			return &pb.CheckFlagResult{Status: true}, nil
		}
		return &pb.CheckFlagResult{Status: false}, nil
	}

	return &pb.CheckFlagResult{}, status.Error(3, "invalid flag format")
}

func (s *server) ServiceRegister(ctx context.Context, in *pb.ServiceRegisterRequest) (*pb.ServiceRegisterResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	username := md.Get("username")[0]
	_, user := getUser(username)
	if in.Key == user.Key {
		flag := fmt.Sprintf("Flag{%s}", generateString(20))
		for i, user := range temporaryUsers {
			if user.Username == username {
				temporaryUsers[i].Services = append(temporaryUsers[i].Services, Service{
					Name: in.Service,
					Flag: flag,
				})
				encryptedFlag, _ := encrypt([]byte(flag), []byte(user.Key))
				return &pb.ServiceRegisterResponse{Flag: encryptedFlag}, nil
			}
		}
	}
	return &pb.ServiceRegisterResponse{}, status.Error(16, "invalid username/key")
}

func generateString(n ...int) (flag string) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if len(n) == 0 {
		n[0] = 10
	}
	if n[0] > 0 {
		s := make([]rune, n[0])
		for i := range s {
			s[i] = letters[rand.Intn(len(letters))]
		}
		return string(s)
	}
	return ""
}

func (s *server) UserRegister(ctx context.Context, in *pb.UserRegisterRequest) (*pb.UserRegisterResponse, error) {
	_, user := getUser(in.Username)
	if user == nil {
		temporaryUsers = append(temporaryUsers, Users{
			Username: in.Username,
			Team:     in.Team,
			Password: in.Password,
		})
		log.Printf("New client: username:%s team:%s", in.Username, in.Team)
		return &pb.UserRegisterResponse{Username: in.Username, Team: in.Team}, nil
	}
	return &pb.UserRegisterResponse{}, status.Errorf(6, "user already exists: %s", in.Username)
}

func (s *server) GetKey(ctx context.Context, in *pb.GetKeyRequest) (*pb.GetKeyResponse, error) {
	for i, user := range temporaryUsers {
		if in.Username == user.Username {
			if in.Password == user.Password {
				key := generateString(32)
				temporaryUsers[i].Key = key
				return &pb.GetKeyResponse{Key: key}, nil
			}
		}
	}
	return &pb.GetKeyResponse{}, status.Error(16, "invalid username/key")
}

func (s *server) ListOfUsers(in *pb.Admin, stream pb.Flags_ListOfUsersServer) error {
	for _, user := range temporaryUsers {
		var services []*pb.Service
		for _, service := range user.Services {
			services = append(services, &pb.Service{Name: service.Name, Status: service.Status})
		}
		stream.Send(&pb.Users{
			Username: user.Username,
			Team:     user.Team,
			Services: services,
		})
	}
	return nil
}

func getUser(username string) (int, *Users) {
	for id, user := range temporaryUsers {
		if username == user.Username {
			return id, &user
		}
	}
	return -1, nil
}

func getService(username string, serviceName string) (int, *Service) {
	_, user := getUser(username)
	for id, service := range user.Services {
		if service.Name == serviceName {
			return id, &service
		}
	}
	return -1, nil
}
func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFlagsServer(s, &server{})
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
