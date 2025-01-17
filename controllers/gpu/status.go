package gpu

import (
	service "github.com/austiecodes/dws/services/gpu"
	"github.com/gin-gonic/gin"
)

// GPUMetricsController
// returns the GPU metrics
func GetStatusController(c *gin.Context) (any, error) {
	data, err := service.NVStatus(c)
	if err != nil {
		return nil, err
	}
	return data, nil
}
