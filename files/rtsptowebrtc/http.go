package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/deepch/vdk/av"

	webrtc "github.com/deepch/vdk/format/webrtcv3"
	"github.com/gin-gonic/gin"
)

type JCodec struct {
	Type string
}

func serveHTTP() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()			// создание роута
	router.Use(CORSMiddleware())

	if _, err := os.Stat("./web"); !os.IsNotExist(err) {
		router.LoadHTMLGlob("web/templates/*")		//загрузка шаблонов
		router.GET("/", HTTPAPIServerIndex)
		//router.GET("/", func(c *gin.Context) {c.HTML(http.StatusOK,"index1.html", gin.H{
		//			"title": "Home Page",
		//		},
		//	)
		//})
		router.GET("/stream/player/:uuid", HTTPAPIServerStreamPlayer)


	}


	router.POST("/stream/receiver/:uuid", HTTPAPIServerStreamWebRTC)
	router.GET("/stream/codec/:uuid", HTTPAPIServerStreamCodec)
	router.POST("/stream", HTTPAPIServerStreamWebRTC2)

	router.GET("/stream/player/:uuid/OpenNautilus/", HTTPOpenNautilus)
	router.GET("/stream/player/:uuid/OnSpotlightURL/", OnSpotlightURL)
	router.GET("/stream/player/:uuid/OffSpotlightURL/", OffSpotlightURL)
	router.GET("/stream/player/:uuid/SetFocusURL/", SetFocusURL)
	router.GET("/stream/player/:uuid/SetFocusUpURL/", SetFocusUpURL)
	router.GET("/stream/player/:uuid/SetFocusDownURL/", SetFocusDownURL)
	router.GET("/stream/player/:uuid/SetZoomUpURL/", SetZoomUpURL)
	router.GET("/stream/player/:uuid/SetZoomDownURL/", SetZoomDownURL)
	router.GET("/stream/player/:uuid/SetZoomURL/", SetZoomURL)
	router.GET("/stream/player/:uuid/MakeScreenshotURL/", MakeScreenshotURL)
	router.GET("/stream/player/:uuid/ZoomOutURL/", ZoomOutURL)
	router.GET("/stream/player/:uuid/ZoomInURL/", ZoomInURL)
	router.GET("/stream/player/:uuid/AlarmoffURL/", AlarmoffURL)
	router.POST("/stream/player/Empiria", ChangeIPAddress)

	//router.GET("/stream/player/:uuid/conf/", ViewIPAddress)

	router.StaticFS("/static", http.Dir("web/static"))
	router.StaticFS("/templates", http.Dir("web/templates"))

	err := router.Run(Config.Server.HTTPPort)
	if err != nil {
		log.Fatalln("Start HTTP Server error", err)
	}
}


func FileNetworkSetting (ip_string string) {

	fmt.Println("FileNetworkSetting ok")
	file, err := os.Create("/lib/systemd/network/10-eth0.network")
	if err != nil {
		return
	}
	defer file.Close()

	var a string
	var countPset int

	for i := 0; i < len(ip_string); i++ {
		if ip_string[i] == '.' {
			countPset ++
			if countPset == 3{
				a = ip_string[:i]
				break
			}
		}
	}

	file.WriteString("[Match]\n" +
		"Name=eth0\n\n" +
		"[Network]\n" +
		"DHCP=ipv4\n" +
		"Address="+ ip_string +"/24\n" +
		"Gateway="+ ip_string +"/24\n" +
		"Broadcast="+ a + ".255\n\n" +
		"[DHCP]\n" +
		"CriticalConnection=true\n")

fmt.Println("FileNetworkSetting ok end")
}


func ViewIPAddress(c *gin.Context){
	data, err := ioutil.ReadFile("config.json") // считали файл в data
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var tmp myConfigST
	err = json.Unmarshal(data, &tmp)	//декодируем файл в tmp
	if err != nil {
		log.Fatalln(err)
	}

	c.JSON(200, tmp.Streams.Empiria.IP)

	log.Printf(color.RedString("отобразить IP %v", tmp.Streams.Empiria.IP))
}



func ChangeIPAddress(c *gin.Context){

	NewIP := c.PostForm("NewIP") // сохраняем параметр comment из POST-запроса
	log.Printf(color.GreenString ("Изменить IP на  %v",NewIP))

	s:= "/usr/bin/ip.sh " + NewIP
	//192.168.3.100
	log.Printf(color.GreenString (s)) //https://github.com/fatih/color


	data, err := ioutil.ReadFile("config.json") // считали файл в data
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var tmp myConfigST
	err = json.Unmarshal(data, &tmp)	//декодируем файл в tmp
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("URL: " + tmp.Streams.Empiria.URL)
	log.Printf(color.GreenString ("OK"))

	//Записывем новый адрес в JSON
	tmp.Streams.Empiria.URL = "rtsp://" + NewIP +":9099/stream"
	tmp.Streams.Empiria.IP = NewIP

	rawDataOut, err := json.MarshalIndent(&tmp, "", "  ")

	log.Printf(color.GreenString ("MarshalIndent OK"))

	err = ioutil.WriteFile("config.json", rawDataOut, 0)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}
	log.Printf(color.GreenString ("WriteFile OK"))

	_, all := Config.list()
	sort.Strings(all)
	//	c.HTML(http.StatusOK, "player.tmpl", gin.H{
	//c.HTML(http.StatusOK, "player.html", gin.H{
	c.HTML(http.StatusOK, "index1.html", gin.H{
		"port":     Config.Server.HTTPPort,
		"suuid":    c.Param("uuid"),
		"suuidMap": all,
		"version":  time.Now().String(),
	})

	go FileNetworkSetting (NewIP)

	lsCmd := exec.Command("sh","/usr/bin/ip.sh",NewIP)
	lsOut, err := lsCmd.Output()
	if err != nil {
		log.Printf(color.RedString("БЕДА! Код ошибки: %v", err))
		return
	}
	fmt.Println(string(lsOut))
	fmt.Println("ChangeIPAddress ok")

	CopyConfig ()

	


}// ChangeIPAddress(c *gin.Context)

func CopyConfig (){
	bytesRead, err := ioutil.ReadFile("config.json")
	if err != nil {
	fmt.Println(err)
	}

	err = ioutil.WriteFile("web/static/conf/conf.json", bytesRead, 0777)
	if err != nil {
	fmt.Println(err)
	}
	log.Printf(color.GreenString ("Copy to web/static/conf/conf.json OK"))
}//func CopyConfig (){


//
//	{
//		"server": {
//		"http_port": ":8083",
//			"ice_servers": ["stun:stun.l.google.com:19302"],
//	"ice_username": "",
//	"ice_credential": ""
//	},
//	"streams": {
//
//	"Empiria": {
//	"on_demand": true,
//	"url": "rtsp://192.168.3.11:9099/stream"
//	}
//	}
//
//}

func WriteSystemIP (){	// запись системного IP-адреса в config.json url:..
	// Получаем все доступные сетевые интерфейсы
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, interf := range interfaces {
	// Список адресов для каждого сетевого интерфейса
		addrs, err := interf.Addrs()
		if err != nil {
			panic(err)
		}

		if interf.Name == "enp7s0" {

			fmt.Printf("Сетевой интерфейс: %s\n", interf.Name)
			for _, add := range addrs {
			if ip, ok := add.(*net.IPNet); ok {
				if ip.IP.To4() !=nil {
						fmt.Printf("\tIP: %v\n", ip.IP)

						var SystemIP = ip.IP.String()

						data, err := ioutil.ReadFile("config.json") // считали файл в data
						if err != nil {
							log.Fatal("Cannot load settings:", err)
						}

						var tmp myConfigST
						err = json.Unmarshal(data, &tmp) //декодируем файл в tmp
						if err != nil {
							log.Fatalln(err)
						}
	
						log.Printf(color.YellowString("Get system IP: " + SystemIP))

						//Записывем новый адрес в JSON
						tmp.Streams.Empiria.URL = "rtsp://" + SystemIP + ":9099/stream"
						tmp.Streams.Empiria.IP = SystemIP
						rawDataOut, err := json.MarshalIndent(&tmp, "", "  ")

						log.Printf(color.YellowString("System IP MarshalIndent OK"))

						err = ioutil.WriteFile("config.json", rawDataOut, 0)
						if err != nil {
							log.Fatal("Cannot write updated settings file:", err)
						}
						log.Printf(color.YellowString("System IP Write OK"))
					}
				}
			}
		}
	}
	CopyConfig () //копируем config.json для отображения на странице
}

func OffSpotlightURL(c *gin.Context){

	fmt.Println("OPEN off-led.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/off-led.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func OnSpotlightURL(c *gin.Context){

	fmt.Println("OPEN on-led.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/on-led.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetFocusURL(c *gin.Context){

	fmt.Println("OPEN startautozoom.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/startautozoom.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetFocusUpURL(c *gin.Context){

	fmt.Println("OPEN startautozoom.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setzoomup.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetFocusDownURL(c *gin.Context){

	fmt.Println("OPEN startautozoom.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setzoomdown.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetZoomURL(c *gin.Context){

	fmt.Println("OPEN startautofocus.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/startautofocus.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetZoomUpURL(c *gin.Context){

	fmt.Println("OPEN startautofocus.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setfocusup.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func SetZoomDownURL(c *gin.Context){

	fmt.Println("OPEN startautofocus.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setfocusdown.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func MakeScreenshotURL(c *gin.Context){

	fmt.Println("OPEN screenshot.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/screenshot.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func ZoomInURL(c *gin.Context){

	fmt.Println("OPEN setfocusdown.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setfocusup.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func ZoomOutURL(c *gin.Context){

	fmt.Println("OPEN setfocusup.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/setfocusdown.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func AlarmoffURL(c *gin.Context){

	fmt.Println("OPEN setfocusup.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/alarmoff.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}

func HTTPOpenNautilus(c *gin.Context){

	fmt.Println("OPEN s.sh")

	c.Header("Cache-Control", "Nautilus")
	fmt.Println("OPEN")
	//lsCmd := exec.Command("bash", "-c", "sh s.sh")
	lsCmd := exec.Command("sh","/usr/bin/s.sh")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN")
	fmt.Println(string(lsOut))

}
//HTTPAPIServerIndex  index
func HTTPAPIServerIndex(c *gin.Context) {
	_, all := Config.list()
	if len(all) > 0 {
		c.Header("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Redirect(http.StatusMovedPermanently, "stream/player/"+all[0])
		fmt.Println(color.BlueString ("HTTPAPIServerIndex _len(all) > 0", all))
	} else {
		//		c.HTML(http.StatusOK, "index.tmpl", gin.H{
				c.HTML(http.StatusOK, "index1.html", gin.H{
			"port":    Config.Server.HTTPPort,
			"version": time.Now().String(),
		})
		fmt.Println(color.RedString ("HTTPAPIServerIndex _else"))
	}

}

//HTTPAPIServerStreamPlayer stream player
func HTTPAPIServerStreamPlayer(c *gin.Context) {
	fmt.Println(color.GreenString ("HTTPAPIServerStreamPlayer:"))
	_, all := Config.list()
	sort.Strings(all)
	//	c.HTML(http.StatusOK, "player.tmpl", gin.H{
		c.HTML(http.StatusOK, "index1.html", gin.H{
		"port":     Config.Server.HTTPPort,
		"suuid":    c.Param("uuid"),
		"suuidMap": all,
		"version":  time.Now().String(),
	})

	//ViewIPAddress(c);
}

//HTTPAPIServerStreamCodec stream codec
func HTTPAPIServerStreamCodec(c *gin.Context) {
	fmt.Println(color.GreenString ("HTTPAPIServerStreamCodec:"))
	if Config.ext(c.Param("uuid")) {
		Config.RunIFNotRun(c.Param("uuid"))
		codecs := Config.coGe(c.Param("uuid"))
		if codecs == nil {
			return
		}
		var tmpCodec []JCodec
		for _, codec := range codecs {
			if codec.Type() != av.H264 && codec.Type() != av.PCM_ALAW && codec.Type() != av.PCM_MULAW && codec.Type() != av.OPUS {
				log.Println("Codec Not Supported WebRTC ignore this track", codec.Type())
				continue
			}
			if codec.Type().IsVideo() {
				tmpCodec = append(tmpCodec, JCodec{Type: "video"})
			} else {
				tmpCodec = append(tmpCodec, JCodec{Type: "audio"})
			}
		}
		b, err := json.Marshal(tmpCodec)
		if err == nil {
			_, err = c.Writer.Write(b)
			if err != nil {
				log.Println("Write Codec Info error", err)
				return
			}
		}
	}
}

//HTTPAPIServerStreamWebRTC stream video over WebRTC
func HTTPAPIServerStreamWebRTC(c *gin.Context) {
	fmt.Println(color.GreenString ("HTTPAPIServerStreamWebRTC:"))
	if !Config.ext(c.PostForm("suuid")) {
		log.Println("Stream Not Found")
		return
	}
	Config.RunIFNotRun(c.PostForm("suuid"))
	codecs := Config.coGe(c.PostForm("suuid"))
	if codecs == nil {
		log.Println("Stream Codec Not Found")
		return
	}
	var AudioOnly bool
	if len(codecs) == 1 && codecs[0].Type().IsAudio() {
		AudioOnly = true
	}
	muxerWebRTC := webrtc.NewMuxer(webrtc.Options{ICEServers: Config.GetICEServers(), ICEUsername: Config.GetICEUsername(), ICECredential: Config.GetICECredential(), PortMin: Config.GetWebRTCPortMin(), PortMax: Config.GetWebRTCPortMax()})
	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"))
	if err != nil {
		log.Println("WriteHeader", err)
		return
	}
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		log.Println("Write", err)
		return
	}
	go func() {
		cid, ch := Config.clAd(c.PostForm("suuid"))
		defer Config.clDe(c.PostForm("suuid"), cid)
		defer muxerWebRTC.Close()
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				log.Println("noVideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame || AudioOnly {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart && !AudioOnly {
					continue
				}
				err = muxerWebRTC.WritePacket(pck)
				if err != nil {
					log.Println("WritePacket", err)
					return
				}
			}
		}
	}()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, x-access-token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

type Response struct {
	Tracks []string `json:"tracks"`
	Sdp64  string   `json:"sdp64"`
}

type ResponseError struct {
	Error  string   `json:"error"`
}

func HTTPAPIServerStreamWebRTC2(c *gin.Context) {
	fmt.Println(color.GreenString ("HTTPAPIServerStreamWebRTC2:"))
	url := c.PostForm("url")
	if _, ok := Config.Streams[url]; !ok {
		Config.Streams[url] = StreamST{
			URL:      url,
			OnDemand: true,
			Cl:       make(map[string]viewer),
		}
	}

	Config.RunIFNotRun(url)

	codecs := Config.coGe(url)
	if codecs == nil {
		log.Println("Stream Codec Not Found")
		c.JSON(500, ResponseError{Error: Config.LastError.Error()})
		return
	}

	muxerWebRTC := webrtc.NewMuxer(
		webrtc.Options{
			ICEServers: Config.GetICEServers(),
			PortMin:    Config.GetWebRTCPortMin(),
			PortMax:    Config.GetWebRTCPortMax(),
		},
	)

	sdp64 := c.PostForm("sdp64")
	answer, err := muxerWebRTC.WriteHeader(codecs, sdp64)
	if err != nil {
		log.Println("Muxer WriteHeader", err)
		c.JSON(500, ResponseError{Error: err.Error()})
		return
	}

	response := Response{
		Sdp64: answer,
	}

	for _, codec := range codecs {
		if codec.Type() != av.H264 &&
			codec.Type() != av.PCM_ALAW &&
			codec.Type() != av.PCM_MULAW &&
			codec.Type() != av.OPUS {
			log.Println("Codec Not Supported WebRTC ignore this track", codec.Type())
			continue
		}
		if codec.Type().IsVideo() {
			response.Tracks = append(response.Tracks, "video")
		} else {
			response.Tracks = append(response.Tracks, "audio")
		}
	}

	c.JSON(200, response)

	AudioOnly := len(codecs) == 1 && codecs[0].Type().IsAudio()

	go func() {
		cid, ch := Config.clAd(url)
		defer Config.clDe(url, cid)
		defer muxerWebRTC.Close()
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				log.Println("noVideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame || AudioOnly {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart && !AudioOnly {
					continue
				}
				err = muxerWebRTC.WritePacket(pck)
				if err != nil {
					log.Println("WritePacket", err)
					return
				}
			}
		}
	}()
}
