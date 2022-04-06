package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	exe, _ := os.Executable()    // 実行ファイルのフルパス
	rootDir := filepath.Dir(exe) // 実行ファイルのあるディレクトリ

	r := gin.Default()
	r.Static("/results", "./results") // 静的ディレクトリとしておかないとHTMLのダウンロードリンクからアクセスできない
	r.LoadHTMLGlob("html/**/*.tmpl")

	// アクセスされたらこれを表示
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
			"title": "t-ocr",
		})
	})

	// uploadされたらこれ
	r.POST("/", func(c *gin.Context) {
		// 時刻オブジェクト
		t := time.Now()
		const layout = "2006-01-02_15-04-05"
		tFormat := t.Format(layout)

		// アップロードされたファイルを格納するディレクトリ作成
		uploadDir := rootDir + "\\" + "uploaded" + "\\" + tFormat
		if err := os.Mkdir(uploadDir, 0777); err != nil {
			fmt.Println(err)
		}

		file, err := c.FormFile("upload")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		log.Println(file.Filename)

		// 特定のディレクトリにファイルをアップロードする
		fileBase := filepath.Base(file.Filename)
		os.Rename(fileBase, tFormat+"_"+fileBase)
		dst := rootDir + "\\uploaded" + "\\" + tFormat + "\\" + fileBase
		log.Println(dst)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Upload file err: %s", err.Error()))
			return
		}

		// アップロードされたzipファイルをunzip
		fmt.Println(uploadDir)
		unzipCmd := exec.Command("7z.exe", "x", "-y", "-o"+uploadDir, dst)
		fmt.Println(unzipCmd)
		if err := unzipCmd.Run(); err != nil {
			fmt.Println("7z unzip command exec error:", err)
		} else {
			fmt.Println("Unzip ok!")
		}

		// アップロードされたzipファイルを削除
		if err := os.Remove(dst); err != nil {
			fmt.Println("Remove error:", err)
		} else {
			fmt.Println("Delete zip file!")
		}

		// OCRする
		if _, err = exec.Command("cmd.exe", "/c", rootDir+"\\"+"t-ocr"+"\\"+"t-ocr.exe", uploadDir, "jpn").CombinedOutput(); err != nil {
			fmt.Println("t-ocr command exec error: ", err)
		} else {
			fmt.Println("cmd.exe", "/c", rootDir+"\\"+"t-ocr"+"\\"+"t-ocr.exe", uploadDir, "jpn")
		}

		// zipする
		dlFile := rootDir + "\\" + "results" + "\\" + tFormat + ".zip"
		if _, err = exec.Command("7z.exe", "a", "-r", "-tzip", dlFile, uploadDir).CombinedOutput(); err != nil {
			fmt.Println("7z zip command exec error: ", err)
		} else {
			fmt.Println("7z.exe", "a", "-r", "-tzip", dlFile, uploadDir)
		}

		// ダウンロードさせるファイル名
		resultFile := tFormat + ".zip"

		// index.tmplを書き換えて、HTMLからダウンロードさせる
		c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
			"title":           "t-ocr",
			"downloadMessage": "Please download: ",
			"downloadfile":    resultFile,
		})
	})

	r.Run(":16")
}
