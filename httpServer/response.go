package httpServer

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
	"time"
)

type ErrorResponse struct {
	Message string `json:"message" example:"status bad request"`
}

func jsonErrorResponse(c *gin.Context, status int, err error) {
	er := ErrorResponse{
		Message: err.Error(),
	}
	c.JSON(status, er)
}

func sendNotModified(c *gin.Context, etag string) (sent bool) {
	ifNonMatch := c.GetHeader("If-None-Match")
	if len(ifNonMatch) < 1 {
		return false
	}

	if strings.TrimPrefix(ifNonMatch, "W/") == etag {
		c.AbortWithStatus(http.StatusNotModified)
		return true
	}

	return false
}

func jsonGetResponse(c *gin.Context, obj interface{}) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	hash := md5.Sum(jsonBytes)
	etag := hex.EncodeToString(hash[:])

	if sendNotModified(c, etag) {
		return
	}

	c.Header("ETag", "W/"+etag)
	c.Data(http.StatusOK, "application/json; charset=utf-8", jsonBytes)
}
func yamlGetResponse(c *gin.Context, obj interface{}) {
	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	hash := md5.Sum(yamlBytes)
	etag := hex.EncodeToString(hash[:])

	if sendNotModified(c, etag) {
		return
	}

	c.Header("ETag", "W/"+etag)
	c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", yamlBytes)
}

func setCacheControlPublic(c *gin.Context, maxAge time.Duration) {
	if maxAge < 0 {
		maxAge = 0
	}
	maxAgeSeconds := int(maxAge.Seconds()) // floor given duration to next lower second
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAgeSeconds))
}
