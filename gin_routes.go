package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func setupGinRoutes(s *Services) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("services", s)
		c.Next()
	})

	r.GET("/health", healthCheck)
	r.GET("/nodes", listNodes)
	r.POST("/job", submitJob)

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func listNodes(c *gin.Context) {
	s := c.MustGet("services").(*Services)
	nodes := s.Cluster.GetNodes()
	c.JSON(http.StatusOK, nodes)
}

func submitJob(c *gin.Context) {
	s := c.MustGet("services").(*Services)

	resultChan := make(chan JobResult, 1)
	s.SubmitJob(Job{
		ID:     fmt.Sprintf("job-%d", time.Now().UnixNano()),
		Data:   []byte("sample data"),
		Result: resultChan,
	})

	result := <-resultChan
	c.JSON(http.StatusOK, gin.H{"result": result})
}
