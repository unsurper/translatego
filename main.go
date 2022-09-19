package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/unicode"
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
	//FileTranslate(viper.GetString("set.path"), viper.GetString("set.output"), &wg)
	Filecompound(viper.GetString("set.path"), viper.GetString("set.compath"), &wg)
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
			wg.Wait()
		}
		if fi.IsDir() {
			FileTranslate(pathname+fi.Name()+"\\", output+fi.Name()+"\\", wg)
		}
	}
	return err
}
func Filecompound(pathname string, output string, wg *sync.WaitGroup) error {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if path.Ext(fi.Name()) == viper.GetString("set.type") {
			Mkdir(output)
			fname := fi.Name()
			wg.Add(1)
			go func() {
				err := ComFile(pathname, fname, output)
				if err != nil {
					fmt.Println("文件:", pathname, ",错误:", err)
				}
				wg.Done()
			}()
			wg.Wait()
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
	fname = strings.Replace(fname, viper.GetString("set.type"), viper.GetString("set.outtype"), 1)
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
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	i := row
	for ; ; i++ {
		a, _, c := read.ReadLine()
		if c == io.EOF {
			break
		}
		if i%row == 0 && len(a) >= 10 {
			if a[3] == 84 && a[7] == 108 {
				context, _ := decoder.Bytes(a[23 : len(a)-2])
				write.WriteString(string(context) + "\n")
				var text []byte
				for {
					a, _, c := read.ReadLine()
					if c == io.EOF {
						break
					}
					if a[1] == 64 {
						if a[3] == 72 && a[5] == 105 {
							context, _ := decoder.Bytes(text)
							write.WriteString("\n" + string(context) + "\n")
							context, _ = decoder.Bytes(a[17 : len(a)-2])
							write.WriteString(string(context) + "\n" + "\n")
							break
						}
					}
					context, _ := decoder.Bytes(a[1 : len(a)-2])
					write.WriteString(string(context))
					text = append(text, a[1:len(a)-2]...)
				}

			}
		} else {
			//write.WriteString(string(a) + "\n")
		}
	}
	write.Flush()
	return
}
func ComFile(pathname string, fname string, output string) (err error) {
	//打开目录ks
	reads, err := os.Open(pathname + fname)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer reads.Close()
	//打开要写入的ks
	writef, err := os.OpenFile(output+fname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	defer writef.Close()
	//打开目录txt
	fname = strings.Replace(fname, viper.GetString("set.type"), viper.GetString("set.outtype"), 1)
	readf, err := os.Open(pathname + fname)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer readf.Close()
	readks := bufio.NewReader(reads)
	readtxt := bufio.NewReader(readf)
	writeks := bufio.NewWriter(writef)
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	b, _, _ := readks.ReadLine()
	writeks.Write(b)
	//找到txt中翻译行
	a, _, c := readtxt.ReadLine()
	for cishu := 1; cishu <= 2; cishu++ {
		a, _, c = readtxt.ReadLine()
		if c == io.EOF {
			break
		}
	}
	var over bool
	for {

		//循环读取ks找到@talk
		for {
			b, _, c := readks.ReadLine()
			if c == io.EOF {
				over = true
				break
			}
			if len(b) > 2 {
				for _, v := range b[1 : len(b)-2] {
					writeks.WriteByte(v)
				}
				writeks.WriteByte(10)
				writeks.WriteByte(0)
			}
			if len(b) > 10 {
				if b[3] == 84 && b[7] == 108 {
					break
				}
			}
		}
		if over == true {
			break
		}
		//写入翻译行
		content, _ := encoder.Bytes(a)
		writeks.Write(content)
		writeks.WriteByte(10)
		writeks.WriteByte(0)
		//更新翻译行
		for cishu := 1; cishu <= 5; cishu++ {
			a, _, c = readtxt.ReadLine()
			if c == io.EOF {
				break
			}
		}

		//更新ks中跳过Hitret
		for {
			b, _, c = readks.ReadLine()
			if c == io.EOF {
				break
			}
			if b[1] == 64 {
				for _, v := range b[1 : len(b)-2] {
					writeks.WriteByte(v)
				}
				writeks.WriteByte(10)
				writeks.WriteByte(0)
				break
			}
		}

	}
	writeks.Flush()
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
