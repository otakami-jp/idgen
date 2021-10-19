package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"os"
	"time"
	"encoding/xml"
	"encoding/json"

	"github.com/discord/fasthttp"
	"github.com/joho/godotenv"
)

var (
	internalAccept 						string = 	"application/json"
	
	epoch 										int = 		1609455600000
	increment 								int = 		0
	internalProcessID 				int = 	  os.Getgid()
	internalProcessWorker 		int = 		0

	incrementBit 							int = 		12
	internalProcessIDBit 			int = 		5
	internalProcessWorkerBit 	int = 		5
)

type AcceptStruct struct {
	Accept string
	Q      float64
}

type ResponseParse struct {
	Accept	[]AcceptStruct
}

type Nekoami struct {
	Id int `xml:"id"`
}

type NekoamiJSON struct {
	Id int `json:"id"`
}

func main() {
	godotenv.Load()


	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8000"
	}

	if err := fasthttp.ListenAndServe(":" + PORT, requestHandler); err != nil {
		panic(err)
	}

	fmt.Println("Server started on port " + PORT)
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	pathReg := regexp.MustCompile(`^\/ids(\/)?`)

	responseParse := ResponseParse {
		Accept: acceptDecoder(ctx),
	}

	switch {
	case pathReg.MatchString(string(ctx.Path())):
		idsHandler(ctx, responseParse)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func acceptDecoder(ctx *fasthttp.RequestCtx) []AcceptStruct {
	accepts := string(ctx.Request.Header.Peek("Accept"))
	
	acceptList := strings.Split(accepts, ", ")

	acceptStructs := []AcceptStruct {}

	for _, a := range acceptList {
		accept := strings.Split(a, ";")

		acceptType := accept[0]

		if len(accept) == 1 {
			accept = append(accept, "q=0.0")
		}

		acceptQRaw := accept[1]

		q, err := strconv.ParseFloat(strings.Split(acceptQRaw, "=")[1], 64)

		if err != nil {
			fmt.Print(err)
		} 

		acceptStructs = append(acceptStructs, AcceptStruct {
			Accept: acceptType,
			Q:      q * 100,
		})
	}

	sort.SliceStable(acceptStructs, func(i, j int) bool {
		return acceptStructs[i].Q > acceptStructs[j].Q
	})
	
	return acceptStructs
}

func idsHandler(ctx *fasthttp.RequestCtx, rp ResponseParse) {
	increment = increment + 1

	id := -1 ^ (increment << incrementBit)
	id = id ^ (internalProcessID << internalProcessIDBit) >> 12
	id = id ^ (internalProcessWorker << internalProcessWorkerBit) >> 17

	timestamp := int(makeTimestamp()) - epoch

	id = (id >> 22) + timestamp

	isJson := true

	for _, a := range rp.Accept {
		if a.Accept == internalAccept {
			isJson = true
			break
		} else if a.Accept == "application/xml" {
			isJson = false
			break
		}
	}

	if isJson {
		sendJson(id, ctx)
	} else {
		sendXml(id, ctx)
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func sendXml(id int, ctx *fasthttp.RequestCtx) {
	xmlResponse, err := xml.Marshal(&Nekoami{ Id: id })

	if err != nil {
		fmt.Print(err)
	}

	ctx.SetContentType("application/xml")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(xmlResponse)
}

func sendJson(id int, ctx *fasthttp.RequestCtx) { 
	jsonResponse, err := json.Marshal(&NekoamiJSON{ Id: id })

	if err != nil {
		fmt.Print(err)
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(jsonResponse)
}