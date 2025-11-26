package main

import "resume-backend-service/internal/config"

func main() {
	app := config.Bootstrap()
	app.Run()
}
