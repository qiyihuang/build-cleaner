package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/joho/godotenv"
	BuildCleaner "github.com/qiyihuang/build-cleaner"
)

const VERSION = "0.1."

func main() {
	if os.Getenv("ENV") != "production" && os.Getenv("ENV") != "test" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}
	functions.HTTP("Clean", BuildCleaner.Clean)
	if err := funcframework.Start("8080"); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
