package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"

	_ "github.com/lib/pq"
)

type config struct {
	host    string
	port    string
	user    string
	pass    string
	dbName  string
	sslMode string
}

func ConnectDb() {

	// loads env file
	configure := &config{
		host:    os.Getenv("DB_HOST"),
		port:    os.Getenv("DB_PORT"),
		user:    os.Getenv("DB_USER"),
		pass:    os.Getenv("DB_PASS"),
		dbName:  os.Getenv("DB_NAME"),
		sslMode: os.Getenv("DB_SSLMODE"),
	}

	psql := fmt.Sprintf("host= %s port= %s user= %s password= %s dbname= %s sslmode= %s",
		configure.host,
		configure.port,
		configure.user,
		configure.pass,
		configure.dbName,
		configure.sslMode)

	fmt.Println("--- DATABASE INITIALIZED ON HOST = " + configure.host + " ---")

	var err error
	models.SqlDb, err = sql.Open("postgres", psql)
	if err != nil {
		log.Fatal("Error connecting to repository - ", err.Error())
	}
}
