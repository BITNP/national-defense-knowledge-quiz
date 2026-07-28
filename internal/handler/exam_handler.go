package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"national-defense-knowledge-quiz/internal/service"
)

type ExamHandler struct {
	svc *service.ExamService
}

func NewExamHandler(svc *service.ExamService) *ExamHandler {
	return &ExamHandler{svc: svc}
}

type startRequest struct {
	Exam      uint   `json:"exam"`
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
}

type submitRequest struct {
	LogID     uint   `json:"log_id"`
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Answers   string `json:"answers"`
}

func (h *ExamHandler) ServeIndex(c *gin.Context) {
	c.File("./static/index.html")
}

func (h *ExamHandler) Info(c *gin.Context) {
	examID := c.Query("exam")
	if examID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "missing exam param"})
		return
	}

	var id uint
	if _, err := fmt.Sscanf(examID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid exam id"})
		return
	}

	info, err := h.svc.Info(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "exam not found"})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *ExamHandler) Start(c *gin.Context) {
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request"})
		return
	}

	if req.StudentID == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "student_id and name required"})
		return
	}

	result, err := h.svc.Start(req.Exam, req.StudentID, req.Name)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ExamHandler) Submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request"})
		return
	}

	if req.StudentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "student_id required"})
		return
	}

	result, err := h.svc.Submit(req.LogID, req.StudentID, req.Name, req.Answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
