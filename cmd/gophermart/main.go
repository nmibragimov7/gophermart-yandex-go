package main

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/handlers"
	"go-musthave-diploma-tpl/internal/jobs"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/router"
	"go-musthave-diploma-tpl/internal/session"
	"log"
	"net/http"
	"time"
)

const (
	logKeyError = "error"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cnf := config.Init()
	sgr := logger.Init()
	defer func() {
		if sgr != nil {
			err := sgr.Sync()
			if err != nil {
				log.Printf("failed to sync logger: %s", err.Error())
			}
		}
	}()
	rps, err := repository.Init(*cnf.DataBase)
	if err != nil {
		sgr.Errorw(
			"failed to init repository",
			"error", err.Error(),
		)

		return fmt.Errorf("failed to init repository: %w", err)
	}
	defer func() {
		err := rps.DB.Close()
		if err != nil {
			sgr.Errorw(
				"failed to close repository connection",
				logKeyError, err.Error(),
			)
		}
	}()

	ssp := &session.SessionProvider{
		Config: cnf,
	}
	hdp := &handlers.HandlerProvider{
		Repository: rps,
		Config:     cnf,
		Sugar:      sgr,
		Session:    ssp,
	}
	rtr := router.RouterProvider{
		Repository: rps,
		Config:     cnf,
		Sugar:      sgr,
		Handler:    hdp,
		Session:    ssp,
	}

	sgr.Error(http.ListenAndServe(*cnf.Server, rtr.Router()))

	jbp := jobs.JobProvider{
		Config:     cnf,
		Sugar:      sgr,
		Repository: rps,
	}
	go jbp.Run(time.Duration(5) * time.Second)
	go jbp.Flush()
	return nil
}
