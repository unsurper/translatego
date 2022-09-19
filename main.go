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
	wg := sync.WaitGroup{}
	InitConfig()
	FileTranslate(viper.GetString("set.path"), viper.GetString("set.output"), &wg)
	wg.Wait()
}

func FileTranslate(pathname string, output string, wg *sync.WaitGroup) error {
	//初始化翻译源
	base := viper.GetString("set.base")
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if path.Ext(fi.Name()) == viper.GetString("set.type") {
			Mkdir(output)
			fname := fi.Name()
			wg.Add(1)
			go func() {
				err := HandleFile(pathname, fname, output, base)
				if err != nil {
					fmt.Println("文件:", pathname, ",错误:", err)
				}
				wg.Done()
			}()
		}
		if fi.IsDir() {
			FileTranslate(pathname+fi.Name()+"\\", output+fi.Name()+"\\", wg)
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

func HandleFile(pathname string, fname string, output string, base string) (err error) {
	reads, err := os.Open(pathname + fname)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer reads.Close()
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
	readsize := bufio.NewReader(reads)
	write := bufio.NewWriter(writef)
	row := viper.GetInt("set.row")
	size := 0
	for ; ; size++ {
		_, _, c := readsize.ReadLine()
		//刷新进度条
		if c == io.EOF {
			break
		}
	}
	//进度条显示
	//var bar bar2.Bar
	//bar.NewOption(0, int64(size+1))
	//过滤头几行
	for i := 0; i < viper.GetInt("set.start"); i++ {
		a, _, c := read.ReadLine()

		if c == io.EOF {
			break
		}
		write.WriteString(string(a) + "\n")
	}
	//跳行翻译
	//decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	for i := row; ; i++ {
		a, _, c := read.ReadLine()
		if c == io.EOF {
			break
		}
		if i%row == 0 && len(a) >= 5 {
			if a[3] == 84 && a[7] == 108 {
				for _, v := range a[1 : len(a)-2] {
					write.WriteByte(v)
				}
				write.WriteByte('\n')
				write.WriteByte(0)
				for {
					a, _, c := read.ReadLine()
					if c == io.EOF {
						break
					}
					if a[3] == 72 {
						write.WriteByte('\n')
						write.WriteByte(0)
						for _, v := range a[1 : len(a)-2] {
							write.WriteByte(v)
						}
						write.WriteByte('\n')
						write.WriteByte(0)
						break
					}
					for _, v := range a[1 : len(a)-2] {
						write.WriteByte(v)
					}
				}

			}
		} else {
			//write.WriteString(string(a) + "\n")
		}
	}
	write.Flush()
	return
}

func Quotes(content string, base string) (string, error) {
	rule, _ := regexp.Compile(`"([^\"]+)"`)
	results := rule.FindAllString(content, -1)
	for _, v := range results {
		s, err := Translate(v, base)
		if err != nil {
			return "", err
		}
		s = strings.Replace(s, "“", "\"", -1)
		s = strings.Replace(s, "”", "\"", -1)
		content = strings.Replace(content, v, s, 1)
	}
	return content, nil
}
func Mkdir(output string) {
	err := os.Mkdir(output, os.ModePerm)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
	}
}
