package main

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strconv"
	"translatego/Baidu"
)

func main() {
	InitConfig()
	base := viper.GetString("set.base")
	S, err := Translate("This is Golang tutorial series.", base)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(S)
	}
}

func Translate(text string, base string) (string, error) {
	count := viper.GetInt(base + ".amount")
	r := rand.Intn(count)
	switch base {
	case "bdpool":
		appid := viper.GetString(base + ".appid" + strconv.Itoa(r))
		appkey := viper.GetString(base + ".key" + strconv.Itoa(r))
		fr := viper.GetString("set.fr")
		to := viper.GetString("set.to")
		S := Baidu.BaiduTranslate(appid, appkey, fr, to, text)
		return S, nil
	default:
		return "", errors.New("not find base")
	}
}

func InitConfig() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
