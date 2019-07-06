package app

import (
	"archive/zip"
	"autossh/src/utils"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Upgrade struct {
	Version string
	latest  map[string]interface{}
}

func showUpgrade() {
	upgrade := Upgrade{Version: Version}
	upgrade.exec()
}

func (upgrade *Upgrade) exec() {

	// 使用协程异步查询最新版本
	var waitGroutp = sync.WaitGroup{}
	islock := true

	go func() {
		utils.Log("正在检测最新版本")
		for {
			if !islock {
				utils.Logln("")
				waitGroutp.Done()
				break
			}
			utils.Log(".")
			time.Sleep(time.Second)
		}
	}()

	go func() {
		upgrade.loadLatestVersion()
		islock = false
		waitGroutp.Done()
	}()

	waitGroutp.Add(2)
	waitGroutp.Wait()

	utils.Logln("当前版本：" + upgrade.Version)
	latestVersion := upgrade.latest["tag_name"].(string)
	ret := upgrade.compareVersion(latestVersion, upgrade.Version)
	if ret <= 0 {
		utils.Logln("感谢您的支持，当前已是最新版本。")
		return
	}

	utils.Logln("检测到新版本：" + latestVersion)
	url := upgrade.downloadUrl()
	if url == "" {
		utils.Errorln("暂不支持" + runtime.GOOS + "系统自动更新，请下载源码包手动编译。")
		return
	}

	filename := path.Base(url)
	savePath := os.TempDir() + filename
	err := upgrade.downloadFile(url, savePath, func(length, downLen int64) {
		process := float64(downLen) / float64(length) * 100

		fmt.Print("\rdownloading " + fmt.Sprintf("%.2f", process) + "%")
	})
	if err != nil {
		utils.Errorln("下载失败：" + err.Error())
		return
	}
	fmt.Print("\rdownloading 100%   \n")

	fullpath, err := upgrade.unzip(savePath, os.TempDir())
	if err != nil {
		utils.Errorln("解压缩失败：" + err.Error())
		return
	}

	cmd := exec.Command(fullpath + "/install")
	output, err := cmd.Output()
	if err != nil {
		utils.Errorln("安装失败")
		return
	}
	utils.Logln(string(output))
}

// 解压缩
func (Upgrade) unzip(zipFile string, destDir string) (string, error) {
	zipReader, err := zip.OpenReader(zipFile)
	fullpath := ""
	if err != nil {
		return fullpath, err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return fullpath, err
			}
		} else {
			fullpath = filepath.Dir(fpath)
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return fullpath, err
			}

			inFile, err := f.Open()
			if err != nil {
				return fullpath, err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fullpath, err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return fullpath, err
			}
		}
	}
	return fullpath, nil
}

// 下载文件
func (Upgrade) downloadFile(url string, downloadPath string, fb func(length, downLen int64)) error {
	var (
		fsize   int64
		buf     = make([]byte, 32*1024)
		written int64
	)
	//创建一个http client
	client := new(http.Client)
	//get方法获取资源
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	//读取服务器返回的文件大小
	fsize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return err
	}
	//创建文件
	file, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if resp.Body == nil {
		return errors.New("body is null")
	}
	defer resp.Body.Close()
	//下面是 io.copyBuffer() 的简化版本
	for {
		//读取bytes
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			//写入bytes
			nw, ew := file.Write(buf[0:nr])
			//数据长度大于0
			if nw > 0 {
				written += int64(nw)
			}
			//写入出错
			if ew != nil {
				err = ew
				break
			}
			//读取是数据长度不等于写入的数据长度
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		//没有错误了快使用 callback

		fb(fsize, written)
	}
	return err
}

// 获取下载地址
func (upgrade *Upgrade) downloadUrl() string {
	sysOS := runtime.GOOS
	if sysOS == "darwin" {
		sysOS = "macOS"
	}

	filename := sysOS + "-" + runtime.GOARCH
	for _, item := range upgrade.latest["assets"].([]interface{}) {
		asset := item.(map[string]interface{})
		if strings.Index(asset["name"].(string), filename) != -1 {
			return asset["browser_download_url"].(string)
		}
	}

	return ""
}

// 读取最新版本信息
func (upgrade *Upgrade) loadLatestVersion() {
	// 使用github api获取最新版本信息
	resp, err := http.Get("https://api.github.com/repos/islenbo/autossh/releases/latest")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var releaseInfo map[string]interface{}
	err = json.Unmarshal(body, &releaseInfo)
	if err != nil {
		panic(err)
	}

	upgrade.latest = releaseInfo
}

// 版本比较
// return int 1: src > other; 0 src == other; -1 src < other
func (Upgrade) compareVersion(src string, other string) int {
	src = strings.Trim(src, "v")
	other = strings.Trim(other, "v")
	v1 := strings.Split(src, ".")
	v2 := strings.Split(other, ".")

	var lim int
	if len(v1) > len(v2) {
		lim = len(v1)
	} else {
		lim = len(v2)
	}

	for {
		if len(v1) >= lim {
			break
		}
		v1 = append(v1, "0")
	}

	for {
		if len(v2) >= lim {
			break
		}
		v2 = append(v2, "0")
	}

	for i := 0; i < lim; i++ {
		num1, _ := strconv.Atoi(v1[i])
		num2, _ := strconv.Atoi(v2[i])

		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}

	return 0
}
