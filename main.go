package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type File struct {
	Name string
	Size int64
	MD5  string
	Path string
}

var justFileContent bool

// 清除文件夹下重复文件
func main() {

	path, err := getPath()

	if err != nil {
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	var filesInfo []File
	wg := sync.WaitGroup{}

	start := time.Now().UnixMilli()
	// 多线程
	go readAllFile(path, &filesInfo, &wg)
	wg.Add(1)
	wg.Wait()

	onlyMap := make(map[string]File)
	//去重
	for _, v := range filesInfo {
		//fmt.Println("读取文件信息:", v)
		name := getName(v)
		f, ok := onlyMap[name]
		if ok {
			fmt.Println("已存在文件:", getName(f), "删除重复文件:", name)
			os.Remove(v.Path)
		} else {
			//fmt.Println("未重复文件:", name)
			onlyMap[name] = v
		}
	}
	fmt.Printf("执行完成！\n执行时间为：%.3f s\n", float64(time.Now().UnixMilli()-start)/1000)

}

func getName(file File) string {
	if justFileContent {
		return file.MD5 + "." + file.Name
	}
	return file.MD5
}

func getFileSize(size int64) string {
	units := []string{"B", "KB", "M", "G", "T"}
	unit := 0
	fSize := float64(size)
	for fSize > 1024 {
		fSize /= 1024
		unit++
	}
	return fmt.Sprintf("%.2f%v", fSize, units[unit])
}

func readAllFile(path string, files *[]File, wg *sync.WaitGroup) {
	curFile, ok := os.Open(path)
	defer curFile.Close()
	defer wg.Done()

	curFileInfo, ok := curFile.Stat()
	if ok != nil {
		fmt.Println("err:", ok)
		return
	}

	if curFileInfo.IsDir() {
		myfiles, _ := curFile.ReadDir(0)
		for _, v := range myfiles {
			go readAllFile(filepath.Join(curFile.Name(), v.Name()), files, wg)
			wg.Add(1)
		}
	} else {
		md5 := countMd5(curFile)
		file := File{
			curFileInfo.Name(),
			curFileInfo.Size(),
			md5,
			filepath.Join(curFile.Name()),
		}

		fmt.Printf("文件读取完成 文件名:%v 文件大小:%v 文件MD5:%v \n", file.Name, getFileSize(file.Size), file.MD5)
		*files = append(*files, file)
	}
}

const bufferSize = 65536

func countMd5(file *os.File) string {
	hash := md5.New()

	for buf, reader := make([]byte, bufferSize), bufio.NewReader(file); ; {
		n, err := reader.Read(buf)

		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err.Error())
		}
		hash.Write(buf[:n])
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getPath() (string, error) {
	var path string

	fmt.Print("请输入扫描地址:")
	fmt.Scanf("%v", &path)

	// 设置清除文件方式
	var t string
	fmt.Printf("是否只删除同名重复文件y/n(y):")
	fmt.Scanf("%v", &t)
	justFileContent = true
	if strings.ToLower(t) == "n" {
		justFileContent = false
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		//fmt.Println(err)
		fmt.Println("文件地址输入有误!")
		return "", err
	}
	if !fileInfo.IsDir() {
		fmt.Println("请输入文件夹地址")
	}
	return path, nil
}
