package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"wm-func/tools/alter-data/models"
	"wm-func/tools/alter-data/services"

	"github.com/gorilla/mux"
)

type APIHandler struct {
	dashboardService *services.DashboardService
}

// NewAPIHandler 创建API处理器实例
func NewAPIHandler() *APIHandler {
	return &APIHandler{
		dashboardService: services.NewDashboardService(),
	}
}

// GetPlatforms 获取所有平台列表
func (h *APIHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	platforms := h.dashboardService.GetAvailablePlatforms()

	fmt.Printf("🔍 DEBUG: API Handler got %d platforms from service\n", len(platforms))
	for i, p := range platforms {
		fmt.Printf("  Handler: %d: %s -> %s\n", i+1, p.Name, p.DisplayName)
	}

	response := models.PlatformResponse{
		Success: true,
		Data:    platforms,
		Message: fmt.Sprintf("Platforms loaded successfully - Count: %d", len(platforms)),
	}

	json.NewEncoder(w).Encode(response)
}

// GetPlatformData 获取指定平台的所有租户数据
func (h *APIHandler) GetPlatformData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	platformName := vars["platform"]

	tenantDataList, err := h.dashboardService.GetPlatformData(platformName)
	if err != nil {
		h.handleError(w, platformName, err, http.StatusBadRequest)
		return
	}

	response := models.DashboardResponse{
		Success:  true,
		Platform: platformName,
		Data:     tenantDataList,
		Message:  "Data loaded successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetTenantData 获取指定租户的平台数据
func (h *APIHandler) GetTenantData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	platformName := vars["platform"]
	tenantIDStr := vars["tenant_id"]

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response := models.DashboardResponse{
			Success:  false,
			Platform: platformName,
			Data:     []models.TenantData{},
			Message:  "Invalid tenant ID: " + tenantIDStr,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	tenantData, err := h.dashboardService.GetTenantData(platformName, tenantID)
	if err != nil {
		h.handleError(w, platformName, err, http.StatusBadRequest)
		return
	}

	response := models.DashboardResponse{
		Success:  true,
		Platform: platformName,
		Data:     []models.TenantData{tenantData},
		Message:  "Data loaded successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// setJSONResponse 设置JSON响应头
func setJSONResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// handleError 统一错误处理
func (h *APIHandler) handleError(w http.ResponseWriter, platform string, err error, statusCode int) {
	response := models.DashboardResponse{
		Success:  false,
		Platform: platform,
		Data:     []models.TenantData{},
		Message:  err.Error(),
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
