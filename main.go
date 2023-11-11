package main

import (
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	database := NewDatabase()

	database.Init()

}
