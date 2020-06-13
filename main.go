package main

import (
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/image/bmp"
)

const (
	spiGetDeskWallpaper = 0x0073
	spiSetDeskWallpaper = 0x0014

	uiParam = 0x0000

	spifUpdateINIFile = 0x01
	spifSendChange    = 0x02
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func convertedWallpaper(bgfile string) string {
	file, err := os.Open(bgfile)
	checkErr(err)
	defer file.Close()

	img, err := jpeg.Decode(file) //解码
	checkErr(err)

	splits := strings.Split(bgfile, `.`)
	bmpPath := splits[0] + `.bmp`
	bmpfile, err := os.Create(bmpPath)
	checkErr(err)
	defer bmpfile.Close()

	err = bmp.Encode(bmpfile, img)
	checkErr(err)
	bmpAbsPath, err := filepath.Abs(bmpPath)
	checkErr(err)
	return bmpAbsPath
}

func setWallPaper(filename string) error {
	filenameUTF16, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}
	systemParametersInfo.Call(
		uintptr(spiSetDeskWallpaper),
		uintptr(uiParam),
		uintptr(unsafe.Pointer(filenameUTF16)),
		uintptr(spifUpdateINIFile|spifSendChange),
	)
	return nil
}

func main() {
	t := time.Now()
	today := t.Format("20060102")
	log.Print(today)
	response, err := http.Get("https://cn.bing.com")
	checkErr(err)
	log.Printf("StatusCode: %d", response.StatusCode)
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		checkErr(err)
		bodyString := string(bodyBytes)
		matchObj := regexp.MustCompile(`(?m)/th\?id=.*?\.jpg`)
		jpgURL := matchObj.FindString(bodyString)
		jpgURL = "https://cn.bing.com" + jpgURL
		filePath := today + ".jpg"
		img, err := os.Create(filePath)
		checkErr(err)
		defer img.Close()
		resp, err := http.Get(jpgURL)
		checkErr(err)
		defer resp.Body.Close()
		b, err := io.Copy(img, resp.Body)
		img.Close()
		checkErr(err)
		log.Println("File size (KB): ", b/1024)
		bmpAbsPath := convertedWallpaper(filePath)
		log.Println("Convert to bmp file: ", bmpAbsPath)
		err = setWallPaper(bmpAbsPath)
		checkErr(err)
		log.Println("Set wallpaper successfully")
		err = os.Remove(bmpAbsPath)
		checkErr(err)
		err = os.Remove(filePath)
		checkErr(err)
	}
}
