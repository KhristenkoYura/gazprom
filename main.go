package main

import (
	"fmt"
	"gazprom/gazprom"
	"gazprom/graceful"
	"gazprom/server"
	"github.com/jmoiron/sqlx"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/francoispqt/onelog"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var err error
	hostname, _ := os.Hostname()

	shutdown := graceful.NewShutdown()
	logger := onelog.New(os.Stderr, onelog.ALL).With(func(e onelog.Entry) {
		// time
		e.String("time", time.Now().Format(time.RFC3339Nano))
		e.String("hostname", hostname)
	})

	logger.Info("start app")
	defer func() {
		if err != nil {
			logger.Fatal(fmt.Sprintf("init app: %s", err))
		}
		logger.Info("stop app")
	}()

	config := &Config{}
	if err = envconfig.Process("", config); err != nil {
		return
	}

	serverHTTP := server.NewServerHTTP(&config.ConfigServer, logger)


	handler := gazprom.Handler{
		DB: NewDB(),
	}
	handler.AddRouter(serverHTTP.V1)

	shutdown.Run(graceful.HTTPServer(serverHTTP.Server()))
	shutdown.Run(graceful.Signal)

	shutdown.Wait(func(err error) {
		logger.Error(err.Error())
	})
}

type Config struct {
	server.ConfigServer
}


func NewDB() *sqlx.DB {
	dsn := `root:root@(127.0.0.1:3306)/gazprom?parseTime=true&timeout=5s&writeTimeout=60s&readTimeout=300s`

	connect, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = connect.Ping()
	if err != nil {
		panic(err)
	}

	return connect
}