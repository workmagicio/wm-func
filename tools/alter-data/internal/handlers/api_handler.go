package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"wm-func/tools/alter-data/internal/service"
	"wm-func/tools/alter-data/models"

	"github.com/gorilla/mux"
)

// APIHandler API请求处理器
type APIHandler struct {
	dashboardService *service.DashboardService
}

// NewAPIHandler 创建API处理器实例
func NewAPIHandler() *APIHandler {
	return &APIHandler{
		dashboardService: service.NewDashboardService(),
	}
}

// GetPlatforms 获取所有平台列表
func (h *APIHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	platforms := h.dashboardService.GetAvailablePlatforms()

	response := models.PlatformResponse{
		Success: true,
		Data:    platforms,
		Message: fmt.Sprintf("成功加载 %d 个平台", len(platforms)),
	}

	json.NewEncoder(w).Encode(response)
}

// GetPlatformData 获取指定平台的所有租户数据
func (h *APIHandler) GetPlatformData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	platformName := vars["platform"]

	// 检查是否强制刷新
	forceRefresh := r.URL.Query().Get("refresh") == "true"

	tenantDataList, err := h.dashboardService.GetPlatformDataWithRefresh(platformName, forceRefresh)
	if err != nil {
		h.handleError(w, platformName, err, http.StatusBadRequest)
		return
	}

	// 获取缓存信息
	cacheInfo := h.dashboardService.GetCacheInfo(platformName)

	message := "数据加载成功"
	if forceRefresh {
		message = "数据已强制刷新"
	} else if cacheInfo != nil {
		message = "数据加载成功（来自缓存）"
	}

	response := models.DashboardResponse{
		Success:   true,
		Platform:  platformName,
		Data:      tenantDataList,
		Message:   message,
		CacheInfo: cacheInfo,
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
			Message:  "无效的租户ID: " + tenantIDStr,
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
		Message:  "数据加载成功",
	}

	json.NewEncoder(w).Encode(response)
}

// setJSONResponse 设置JSON响应头
func setJSONResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// RefreshPlatformData 刷新指定平台的缓存数据
func (h *APIHandler) RefreshPlatformData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	platformName := vars["platform"]

	err := h.dashboardService.RefreshPlatformCache(platformName)
	if err != nil {
		h.handleError(w, platformName, err, http.StatusBadRequest)
		return
	}

	// 获取刷新后的数据
	tenantDataList, err := h.dashboardService.GetPlatformData(platformName)
	if err != nil {
		h.handleError(w, platformName, err, http.StatusBadRequest)
		return
	}

	// 获取缓存信息
	cacheInfo := h.dashboardService.GetCacheInfo(platformName)

	response := models.DashboardResponse{
		Success:   true,
		Platform:  platformName,
		Data:      tenantDataList,
		Message:   "缓存已刷新",
		CacheInfo: cacheInfo,
	}

	json.NewEncoder(w).Encode(response)
}

// GetCacheStats 获取缓存统计信息
func (h *APIHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	stats := h.dashboardService.GetCacheStats()

	response := models.APIResponse{
		Success: true,
		Message: "缓存统计信息获取成功",
		Data:    stats,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTenants 获取租户列表
func (h *APIHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	tenantList, err := h.dashboardService.GetTenantList()
	if err != nil {
		response := models.TenantListResponse{
			Success: false,
			Data:    []models.TenantInfo{},
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.TenantListResponse{
		Success: true,
		Data:    tenantList,
		Message: fmt.Sprintf("成功加载 %d 个租户", len(tenantList)),
	}

	json.NewEncoder(w).Encode(response)
}

// GetRecentTenants 获取最近注册的租户列表
func (h *APIHandler) GetRecentTenants(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	// 检查是否强制刷新
	forceRefresh := r.URL.Query().Get("refresh") == "true"

	recentTenants, err := h.dashboardService.GetRecentRegisteredTenantsWithRefresh(forceRefresh)
	if err != nil {
		response := models.RecentTenantsResponse{
			Success: false,
			Data:    []models.TenantInfo{},
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	message := fmt.Sprintf("成功获取 %d 个最近注册租户", len(recentTenants))
	if forceRefresh {
		message = fmt.Sprintf("已刷新最近注册租户列表 (%d 个)", len(recentTenants))
	}

	response := models.RecentTenantsResponse{
		Success: true,
		Data:    recentTenants,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// GetFrequentTenants 获取经常访问的租户列表
func (h *APIHandler) GetFrequentTenants(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	// 检查是否强制刷新
	forceRefresh := r.URL.Query().Get("refresh") == "true"

	frequentTenants, err := h.dashboardService.GetFrequentTenants()
	if err != nil {
		response := models.FrequentTenantsResponse{
			Success: false,
			Data:    []models.TenantAccessRecord{},
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	message := fmt.Sprintf("成功获取 %d 个经常访问租户", len(frequentTenants))
	if forceRefresh {
		message = fmt.Sprintf("已刷新经常访问租户列表 (%d 个)", len(frequentTenants))
	}

	response := models.FrequentTenantsResponse{
		Success: true,
		Data:    frequentTenants,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTenantCrossPlatformData 获取指定租户的跨平台数据
func (h *APIHandler) GetTenantCrossPlatformData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response := models.TenantCrossPlatformResponse{
			Success:    false,
			TenantID:   0,
			TenantName: "",
			Data:       models.CrossPlatformData{},
			Message:    "无效的租户ID: " + tenantIDStr,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 检查是否强制刷新
	forceRefresh := r.URL.Query().Get("refresh") == "true"

	crossPlatformData, err := h.dashboardService.GetTenantCrossPlatformDataWithRefresh(tenantID, forceRefresh)
	if err != nil {
		response := models.TenantCrossPlatformResponse{
			Success:    false,
			TenantID:   tenantID,
			TenantName: fmt.Sprintf("Tenant %d", tenantID),
			Data:       models.CrossPlatformData{},
			Message:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取缓存信息
	cacheInfo := h.dashboardService.GetTenantCacheInfo(tenantID)

	message := "数据加载成功"
	if forceRefresh {
		message = "数据已强制刷新"
	} else if cacheInfo != nil {
		message = "数据加载成功（来自缓存）"
	}

	response := models.TenantCrossPlatformResponse{
		Success:    true,
		TenantID:   tenantID,
		TenantName: crossPlatformData.TenantName,
		Data:       crossPlatformData,
		Message:    message,
		CacheInfo:  cacheInfo,
	}

	json.NewEncoder(w).Encode(response)
}

// RefreshTenantData 刷新指定租户的缓存数据
func (h *APIHandler) RefreshTenantData(w http.ResponseWriter, r *http.Request) {
	setJSONResponse(w)

	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response := models.TenantCrossPlatformResponse{
			Success:    false,
			TenantID:   0,
			TenantName: "",
			Data:       models.CrossPlatformData{},
			Message:    "无效的租户ID: " + tenantIDStr,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = h.dashboardService.RefreshTenantCache(tenantID)
	if err != nil {
		response := models.TenantCrossPlatformResponse{
			Success:    false,
			TenantID:   tenantID,
			TenantName: fmt.Sprintf("Tenant %d", tenantID),
			Data:       models.CrossPlatformData{},
			Message:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取刷新后的数据
	crossPlatformData, err := h.dashboardService.GetTenantCrossPlatformData(tenantID)
	if err != nil {
		response := models.TenantCrossPlatformResponse{
			Success:    false,
			TenantID:   tenantID,
			TenantName: fmt.Sprintf("Tenant %d", tenantID),
			Data:       models.CrossPlatformData{},
			Message:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取缓存信息
	cacheInfo := h.dashboardService.GetTenantCacheInfo(tenantID)

	response := models.TenantCrossPlatformResponse{
		Success:    true,
		TenantID:   tenantID,
		TenantName: crossPlatformData.TenantName,
		Data:       crossPlatformData,
		Message:    "缓存已刷新",
		CacheInfo:  cacheInfo,
	}

	json.NewEncoder(w).Encode(response)
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
