package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/darkMoon1973/PortScanner/common/util"
	"github.com/gin-gonic/gin"
)

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func httpResp(c *gin.Context, httpCode int, msg string, data interface{}) {
	c.JSON(httpCode, response{
		Code:    httpCode,
		Message: msg,
		Data:    data,
	})
}

func newHttpServer(httpPort int) {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Status(http.StatusOK)
		}
	})

	agentGroup := r.Group("/api", agentAuth())
	agentGroup.GET("/download", httpAgentDownload)
	agentGroup.GET("/register", httpAgentRegister)
	err := r.Run(":" + strconv.Itoa(httpPort))

	time.Sleep(time.Second)
	if err != nil {
		logic.Log.Error("start http server error:", err)
	}
}

func agentAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		logic.Conf.Http.Token: logic.Conf.Http.Token,
	})
}

func httpAgentDownload(c *gin.Context) {
	s := util.ReadFile("scanAgent")
	fileName := []byte(s)
	c.Status(http.StatusOK)
	c.Header("Content-Length", strconv.Itoa(len(fileName)))
	_, err := c.Writer.Write(fileName)
	if err != nil {
		logic.Log.Error("gin write agent failed, ", err)
	}
}

func httpAgentRegister(c *gin.Context) {
	resp := map[string]interface{}{
		"redisUrl":      logic.Conf.RedisUrl,
		"autoInterface": logic.Conf.Masscan.Enable,
	}
	httpResp(c, http.StatusOK, "success", resp)
}
