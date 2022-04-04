package routes

import (
	"example.com/controllers"

	"github.com/gin-gonic/gin"
)

func Init() (route *gin.Engine) {

	gin.SetMode(gin.ReleaseMode) //TODO: set release mode for production
	router := gin.New()
	v1 := router.Group("/v1")
	blocks := v1.Group("/blocks")
	blocks.GET("/", controllers.GetBlocks)
	blocks.GET("/:num", controllers.GetBlock)

	v1.GET("/transaction/:txHash", controllers.GetTransaction)
	return router

}
