package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	Server    *string
	Accrual   *string
	DataBase  *string
	SecretKey *string
}

func Init() *Config {
	instance := Config{
		Server:    nil,
		Accrual:   nil,
		DataBase:  nil,
		SecretKey: nil,
	}

	flags := flag.NewFlagSet("config", flag.ContinueOnError)

	instance.Server = flags.String("a", ":4200", "Server address")
	instance.Accrual = flags.String("r", "http://localhost:8080", "Accrual address")
	instance.SecretKey = flags.String("s", "secret_key", "JWT secret key")
	instance.DataBase = flags.String(
		"d",
		"",
		"Database URL",
	) // host=localhost user=postgres password=admin dbname=gophermart sslmode=disable

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Printf("failed to parse flags: %s", err.Error())
	}

	if envServerAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		instance.Server = &envServerAddress
	}
	if envDataBaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		instance.DataBase = &envDataBaseURI
	}
	if envDatabase, ok := os.LookupEnv("SECRET_KEY"); ok {
		instance.SecretKey = &envDatabase
	}
	if envAccrualAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		instance.Accrual = &envAccrualAddress
	}

	return &instance
}
