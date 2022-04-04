package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/joho/godotenv"
	_ "github.com/qiyihuang/build-cleaner"
)

func main() {
	if os.Getenv("ENV") != "production" && os.Getenv("ENV") != "test" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}
	port := "8080"
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
