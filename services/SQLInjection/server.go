package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"errors"
	"fmt"
	pb "github.com/Ivanhahanov/grpcFlags/flags"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"time"
)

var flag string

const address = "192.168.31.66:50051"

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

func getFlag() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewFlagsClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, "username", "user")
	defer cancel()
	r, err := c.GetKey(ctx, &pb.GetKeyRequest{
		Username: "user",
		Password: "userSuperSecretKey",
	})
	if err != nil {
		log.Fatalf("sorry, %v", err)
	}
	secretKey := r.Key

	reg, err := c.ServiceRegister(ctx, &pb.ServiceRegisterRequest{
		Service: "sql",
		Key:     secretKey,
	})
	if err != nil {
		log.Fatalf("could not send request: %v", err)
	}
	decryptedFlag, _ := decrypt(reg.Flag, []byte(secretKey))
	flag = string(decryptedFlag)
}

func initDb() {
	db, err := sql.Open("sqlite3", "user.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE users " +
		"(`_id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"`username` VARCHAR(64), " +
		"`password` VARCHAR(64))")
	if err != nil {
		panic(err)
	}
	db.Exec("insert into users (username, password) values ('admin', 'secret')")

	fmt.Println("Creating user database")

}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	var form = `
	<div><form action="/login" method="POST">
	<p>Username: <input type="text" name="username" /></p>
	<p>Password: <input type="password" name="password" /></p>
	<p><input type="submit" value="Login" /></p>
	</form></div>`
	fmt.Fprint(w, form)
}

func Login(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "user.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	login := r.FormValue("username")
	password := r.FormValue("password")
	sqlRequest := "SELECT username from users where username='" + login + "' and password='" + password + "'"
	fmt.Println(sqlRequest)
	rows, err := db.Query(sqlRequest)
	if err != nil {
		panic(err)
	}
	var username string
	rows.Next()
	rows.Scan(&username)
	fmt.Println(username)
	if username == "admin" {
		fmt.Fprint(w, flag)
	} else {
		fmt.Fprint(w, "Auth Error")
	}
}

func main() {
	initDb()
	getFlag()
	fmt.Println("Starting Web Server with Open Redirect vulnerability")
	http.HandleFunc("/login", Login)
	http.HandleFunc("/", renderIndex)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
