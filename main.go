package main

import (
	"bytes"
	"context"
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
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
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
	router.StaticFile("/", "./static/index.html")
	//router.StaticFile("/jquery.js", "./static/jquery.js")
	router.Static("/assets", "./static/assets")
	//router.Static("/css", "./static/css")
	os.MkdirAll("./download", fs.ModePerm)
	router.StaticFS("/download", gin.Dir("./download", true))
	os.MkdirAll("./upload", fs.ModePerm)
}

var shutdown = make(chan bool)

func StartServer(addr string, port int64) {
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: router,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

	}()
	go func() {
		if <-shutdown {
			println("Stoping Server")
			ctx := context.Background()
			err := srv.Shutdown(ctx)
			if err != nil {
				println(err)
				return
			}
			println("Starting Server")
			StartServer(addr, port)
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
	ipaddr := ""
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if ipaddr == "" {
					ipaddr = ipnet.IP.To4().String()
				}
				println(ipnet.IP.To4().String())
			}
		}

	}
	port := 10000 + rand.Int63n(15535)
	url := fmt.Sprintf("%s:%d", ipaddr, port)
	StartServer(ipaddr, port)

	a := app.New()
	w := a.NewWindow("Copy And Past")
	QrcodeImageData, err := qrcode.Encode(fmt.Sprintf("http://%s", url), qrcode.Medium, 256)
	QrcodeImage, _, err := image.Decode(bytes.NewReader(QrcodeImageData))
	if err != nil {
		println(err)
		return
	}
	imageCanvas := canvas.NewImageFromImage(QrcodeImage)
	imageCanvas.FillMode = canvas.ImageFillOriginal
	w.SetContent(container.NewVBox(
		widget.NewLabel(fmt.Sprintf("http://%s", url)),
		imageCanvas,
		widget.NewButton("Stop", func() {
			shutdown <- true

		}),
	))
	w.ShowAndRun()
}
