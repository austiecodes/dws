package gpu

import (
	"net/http"

	service "github.com/austiecodes/dws/services/gpu"
	"github.com/gin-gonic/gin"
)

// GPUMetrics
// returns the GPU metrics
func GetGPUStatus(c *gin.Context) {
	data, err := service.NVStatus(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, data)
}
