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
