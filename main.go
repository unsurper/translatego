package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"translatego/Baidu"
)

func main() {
	InitConfig()
	FileTranslate(viper.GetString("set.path"), viper.GetString("set.output"))
}

func FileTranslate(pathname string, output string) error {
	//初始化翻译源
	var wg sync.WaitGroup
	base := viper.GetString("set.base")
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if path.Ext(fi.Name()) == viper.GetString("set.type") {
			Mkdir(output)
			fname := fi.Name()
			wg.Add(1)
			go func() {
				HandleFile(pathname, fname, output, base)
				wg.Done()
			}()
			wg.Wait()
		}
		if fi.IsDir() {
			FileTranslate(pathname+fi.Name()+"\\", output+fi.Name()+"\\")
		}
	}
	return err
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

func HandleFile(pathname string, fname string, output string, base string) {
	readf, err := os.Open(pathname + fname)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer readf.Close()

	writef, err := os.OpenFile(output+fname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	defer writef.Close()

	read := bufio.NewReader(readf)
	write := bufio.NewWriter(writef)
	row := viper.GetInt("set.row")
	//过滤头几行
	for i := 0; i < viper.GetInt("set.start"); i++ {
		a, _, c := read.ReadLine()
		if c == io.EOF {
			break
		}
		write.WriteString(string(a) + "\n")
	}
	//跳行翻译
	for i := row; ; i++ {
		a, _, c := read.ReadLine()
		if c == io.EOF {
			break
		}

		if i%row == 0 {
			content := string(a)
			content = Quotes(content, base)
			write.WriteString(content + "\n")
		} else {
			write.WriteString(string(a) + "\n")
		}
	}
	write.Flush()

}
func Quotes(content string, base string) string {
	rule, _ := regexp.Compile(`"([^\"]+)"`)
	results := rule.FindAllString(content, -1)
	for _, v := range results {
		s, _ := Translate(v, base)
		s = strings.Replace(s, "“", "\"", -1)
		s = strings.Replace(s, "”", "\"", -1)
		content = strings.Replace(content, v, s, 1)
	}
	return content
}
func Mkdir(output string) {
	err := os.Mkdir(output, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}
