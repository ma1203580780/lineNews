package controller

import (
	"lineNews/model"
	"net/http"

	"lineNews/agent/logutil"

	"github.com/gin-gonic/gin"
)

// HandleBaikeSearch 处理百度百科搜索请求
func HandleBaikeSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "keyword 参数不能为空",
		})
		return
	}

	logutil.LogInfo("百度百科搜索请求: %s", keyword)

	// 调用 model 层
	response, err := model.BaiduBaikeSearchSimple(keyword)
	if err != nil {
		logutil.LogError("百科搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "百科搜索失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"keyword": keyword,
		"data":    response,
	})
}

// HandleBaikeSearchByLemmaId 通过词条ID搜索百度百科
func HandleBaikeSearchByLemmaId(c *gin.Context) {
	lemmaId := c.Query("lemma_id")
	if lemmaId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lemma_id 参数不能为空",
		})
		return
	}

	logutil.LogInfo("百度百科词条ID搜索: %s", lemmaId)

	// 调用 model 层
	response, err := model.BaiduBaikeSearchByLemmaId(lemmaId)
	if err != nil {
		logutil.LogError("百科词条ID搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "百科搜索失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"lemma_id": lemmaId,
		"data":     response,
	})
}
