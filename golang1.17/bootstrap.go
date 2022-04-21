package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	handler := getEnv("AWS_LAMBDA_FUNCTION_HANDLER", getEnv("_HANDLER", "handler"))

	cmd := exec.Command("/var/task/" + handler)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
