package handlers

import (
	"net/http"
	"path/filepath"
)

// ServeStatic 提供静态文件服务
func ServeStatic() http.Handler {
	staticDir := "static"
	fileServer := http.FileServer(http.Dir(staticDir))
	return http.StripPrefix("/static/", fileServer)
}

// ServeIndex 提供主页面
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join("static", "index.html")
	http.ServeFile(w, r, indexPath)
}

// ServeAttributionPage 提供归因订单分析页面
func ServeAttributionPage(w http.ResponseWriter, r *http.Request) {
	attributionPath := filepath.Join("static", "attribution.html")
	http.ServeFile(w, r, attributionPath)
}

// ServeAmazonOrdersPage 提供Amazon订单分析页面
func ServeAmazonOrdersPage(w http.ResponseWriter, r *http.Request) {
	amazonOrdersPath := filepath.Join("static", "amazon-orders.html")
	http.ServeFile(w, r, amazonOrdersPath)
}

// ServeFairingPage 提供Fairing分析页面
func ServeFairingPage(w http.ResponseWriter, r *http.Request) {
	fairingPath := filepath.Join("static", "fairing.html")
	http.ServeFile(w, r, fairingPath)
}
