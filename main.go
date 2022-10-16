package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"image"
	"io/fs"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"syncClipBoard/utils"
	"time"
)

type postStringBody struct {
	Message string `json:"message"`
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var socketSlice []*websocket.Conn

//go:embed static/assets/* static/index.html
var staticFile embed.FS

//go:embed static/index.html
var indexFile string

func ping(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	err = ws.WriteJSON(gin.H{"message": newS})
	if err != nil {
		fmt.Println(err)
		return
	}
	socketSlice = append(socketSlice, ws)
}

func sendToAllClient(data string) {
	for i, conn := range socketSlice {
		err := conn.WriteJSON(gin.H{"Message": data})
		if err != nil {
			fmt.Println(err)
			socketSlice = append(socketSlice[:i], socketSlice[i+1:]...)
			return
		}
	}
}

var (
	oldS = ""
	newS = ""
)

func LoopCheckClipBoard() {

	var err error
	for true {
		time.Sleep(1)
		newS, err = clipboard.ReadAll()

		if err != nil {
			println(err)
		}
		if newS != oldS {
			//fmt.Println(newS)
			sendToAllClient(newS)
			oldS = newS
		}
	}

}

var router = gin.Default()

func init() {
	//gin.SetMode(gin.ReleaseMode)
	router.POST("/set", func(context *gin.Context) {
		body := postStringBody{}
		err := context.BindJSON(&body)
		if err != nil {
			context.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}
		err = clipboard.WriteAll(body.Message)
		oldS = body.Message
		newS = body.Message
		if err != nil {
			context.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}

	})
	router.GET("/get", func(context *gin.Context) {
		all, err := clipboard.ReadAll()
		if err != nil {
			context.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}
		context.JSON(200, gin.H{"Message": all})
	})
	router.GET("/ws", ping)
	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("upload")
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}
		err = c.SaveUploadedFile(file, "./upload/"+file.Filename)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}

		cPath, err := os.Getwd()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}
		utils.OpenDir(path.Join(cPath, "upload"))
		c.JSON(200, gin.H{"Message": "success"})
	})
	router.GET("/", func(c *gin.Context) {
		c.Header("content-type", "text/html;charset=utf-8")
		c.String(200, indexFile)
	})
	router.StaticFS("/s", http.FS(staticFile))
	router.Any("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")

		c.FileFromFS("static/assets"+filepath, http.FS(staticFile))
	})
	os.MkdirAll("./download", fs.ModePerm)
	router.GET("/download", func(c *gin.Context) {
		queryPath := c.Query("path")
		pathArray := strings.Split(queryPath, "/")[2:]
		pathArray = append([]string{".", "download"}, pathArray...)
		realPath := path.Join(pathArray...)
		fileName := pathArray[len(pathArray)-1]
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		println(realPath)
		c.File(realPath)
	})

	os.MkdirAll("./upload", fs.ModePerm)
	router.GET("/filelist", func(c *gin.Context) {
		type fileType struct {
			FileName string `json:"fileName"`
			IsDir    bool   `json:"isDir"`
		}
		queryPath := c.Query("path")
		pathArray := strings.Split(queryPath, "/")[1:]
		pathArray = append([]string{".", "download"}, pathArray...)
		realPath := path.Join(pathArray...)
		files, err := os.ReadDir(realPath)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"Message": err})
			return
		}
		r := make([]fileType, 0)
		for _, file := range files {
			r = append(r, fileType{
				FileName: file.Name(),
				IsDir:    file.IsDir(),
			})
		}
		c.JSON(200, r)
	})
}

var shutdown = make(chan bool)
var shutdownFinished = make(chan bool)

var srv = &http.Server{}

func StartServer(addr string, port int64) {
	srv = &http.Server{Addr: fmt.Sprintf("%s:%d", addr, port), Handler: router}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			println("listen: %s", err)
		}

	}()

}

func main() {
	go LoopCheckClipBoard()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		println(err)
		return
	}
	ipaddrSlice := []string{}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipaddrSlice = append(ipaddrSlice, ipnet.IP.To4().String())
			}
		}
	}

	port := 10000 + rand.Int63n(15535)
	url := ""

	a := app.New()
	w := a.NewWindow("Copy And Past")
	QrcodeImage := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imageCanvas := canvas.NewImageFromImage(QrcodeImage)
	imageCanvas.FillMode = canvas.ImageFillOriginal

	rebootButton := widget.NewButton("Stop", func() {
		err = srv.Shutdown(context.Background())
		if err != nil {
			println(err)
			return
		}
	})
	w.SetContent(container.NewVBox(
		widget.NewSelect(ipaddrSlice, func(s string) {
			c := context.Background()
			err := srv.Shutdown(c)
			if err != nil {
				println(err)
				return
			}
			url = fmt.Sprintf("http://%s:%d", s, port)
			QrcodeImageData, err := qrcode.Encode(url, qrcode.Medium, 256)
			QrcodeImage, _, err := image.Decode(bytes.NewReader(QrcodeImageData))
			if err != nil {
				println(err)
				return
			}
			imageCanvas.Image = QrcodeImage
			imageCanvas.Refresh()
			StartServer(s, port)
		}),
		imageCanvas,
		rebootButton,
	))
	w.ShowAndRun()
}
