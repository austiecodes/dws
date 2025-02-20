package gpu

import (
	"net/http"

	services "github.com/austiecodes/dws/services/gpu"
	"github.com/gin-gonic/gin"
)

// GPUMetrics
// returns the GPU metrics
func GetGPUStatus(c *gin.Context) {
	data, err := services.GPUStatus(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, data)
}
