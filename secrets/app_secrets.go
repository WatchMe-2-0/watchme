package secrets

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.AddConfigPath("./secrets")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	file, err := os.Open("./secrets/app.env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = viper.ReadConfig(file)
	if err != nil {
		log.Fatal(err)
	}
}
