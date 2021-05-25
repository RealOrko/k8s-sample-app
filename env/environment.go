package env

import "C"
import (
	"log"
	"os"
	"strconv"
)

// *** Environment ***

type Environment struct {
	HttpPort        string
	FailHealthCheck bool
	RequestDelay    int
}

func getEnvironmentKeyStr(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvironmentKeyBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		castValue, err := strconv.ParseBool(value)
		if err == nil {
			return castValue
		} else {
			log.Fatal(err)
		}
	}
	return fallback
}

func getEnvironmentKeyInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		castValue, err := strconv.Atoi(value)
		if err == nil {
			return castValue
		} else {
			log.Fatal(err)
		}
	}
	return fallback
}

func GetEnvironment() Environment {
	env := Environment{
		HttpPort:        getEnvironmentKeyStr("PORT", "8000"),
		FailHealthCheck: getEnvironmentKeyBool("FAIL_HEALTH_CHECK", false),
		RequestDelay:    getEnvironmentKeyInt("REQUEST_DELAY", 5),
	}
	log.Printf("env:PORT: %s\n", env.HttpPort)
	log.Printf("env:REQUEST_DELAY: %v\n", env.RequestDelay)
	log.Printf("env:FAIL_HEALTH_CHECK: %v\n", env.FailHealthCheck)
	return env
}


