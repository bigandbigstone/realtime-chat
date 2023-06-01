package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var roomManager *Manager

func main() {
	roomManager = NewRoomManager()
	router := gin.Default()
	router.SetHTMLTemplate(html)

	router.GET("/room/:roomid", roomGET)
	router.POST("/room/:roomid", roomPOST)
	router.DELETE("/room/:roomid", roomDELETE)
	router.GET("/stream/:roomid", stream)
	router.GET("/msgstream/:roomid", roomStreamGET)
	router.GET("/delstream/:roomid", roomStreamDEL)

	router.Run(":8080")
}

func stream(c *gin.Context) {
	roomid := c.Param("roomid")
	listener := roomManager.OpenListener(roomid)
	defer roomManager.CloseListener(roomid, listener)

	clientGone := c.Request.Context().Done()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case message := <-listener:
			c.SSEvent("message", message)
			return true
		}
	})
}

func roomStreamGET(c *gin.Context) {
	roomid := c.Param("roomid")
	messagelist := roomManager.GetMessageCache(roomid)

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"roomid":       roomid,
		"messagecache": messagelist,
	})

	// to do，聊天室的聊天策略可能有问题
	// 貌似优雅的清空策略，但可能产生数据冲突问题
	// roomManager.messagecache[roomid] = roomManager.messagecache[roomid][0:0]
}

func roomStreamDEL(c *gin.Context) {
	roomid := c.Param("roomid")
	// roomManager messageCache 清空
	roomManager.DelMessageCache(roomid)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"roomid": roomid,
	})

	// to do，聊天室的聊天策略可能有问题
	// 貌似优雅的清空策略，但可能产生数据冲突问题
	// roomManager.messagecache[roomid] = roomManager.messagecache[roomid][0:0]
}

func roomGET(c *gin.Context) {
	roomid := c.Param("roomid")
	userid := "username"
	c.HTML(http.StatusOK, "chat_room", gin.H{
		"roomid": roomid,
		"userid": userid,
	})
}

func roomPOST(c *gin.Context) {
	roomid := c.Param("roomid")
	userid := c.PostForm("user")
	message := c.PostForm("message")
	roomManager.Submit(userid, roomid, message)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
	})
}

func roomDELETE(c *gin.Context) {
	roomid := c.Param("roomid")
	roomManager.DeleteBroadcast(roomid)
}
