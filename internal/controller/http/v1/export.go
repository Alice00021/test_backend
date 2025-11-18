package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"test_go/internal/controller/http/errors"
	"test_go/internal/usecase"
	"test_go/pkg/logger"
)

type exportRoutes struct {
	l  logger.Interface
	uc usecase.Export
}

func NewExportRoutes(privateGroup *gin.RouterGroup, l logger.Interface, uc usecase.Export) {
	r := &exportRoutes{l, uc}
	{
		h := privateGroup.Group("/export")
		h.GET("/statistics", r.generateExportFile)
		h.GET("/commands/csv", r.exportCommandsToCSV)
		h.GET("/commands/pdf", r.exportCommandsToPDF)
		h.GET("/operations/csv", r.exportOperationsToCSV)
		h.GET("/operations/pdf", r.exportOperationsToPDF)
	}
}

func (r *exportRoutes) generateExportFile(c *gin.Context) {
	file, err := r.uc.GenerateExcelFile(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - generateExportFile")
		errors.ErrorResponse(c, err)
		return
	}
	defer file.Close()

	fileName, err := r.uc.SaveToFile(file)
	if err != nil {
		r.l.Error(err, "http - v1 - generateExportFile")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, fileName)
}

func (r *exportRoutes) exportCommandsToCSV(c *gin.Context) {
	res, fileName, err := r.uc.ExportCommandsToCSV(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - exportCommandsToCSV")
		errors.ErrorResponse(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "text/csv")

	c.Data(http.StatusOK, "text/csv", res)
}

func (r *exportRoutes) exportCommandsToPDF(c *gin.Context) {
	res, fileName, err := r.uc.ExportCommandsToPDF(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - exportCommandsToPDF")
		errors.ErrorResponse(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/pdf")

	c.Data(http.StatusOK, "application/pdf", res)
}

func (r *exportRoutes) exportOperationsToCSV(c *gin.Context) {
	res, fileName, err := r.uc.ExportOperationsToCSV(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - exportOperationsToCSV")
		errors.ErrorResponse(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "text/csv")

	c.Data(http.StatusOK, "text/csv", res)
}

func (r *exportRoutes) exportOperationsToPDF(c *gin.Context) {
	res, fileName, err := r.uc.ExportOperationsToPDF(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - exportOperationsToPDF")
		errors.ErrorResponse(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/pdf")

	c.Data(http.StatusOK, "application/pdf", res)
}
