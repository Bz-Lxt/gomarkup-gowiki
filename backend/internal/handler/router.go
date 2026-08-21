package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/collab/hub"
	"gowiki/internal/middleware"
	"gowiki/internal/service"
)

type Deps struct {
	Auth      *AuthHandler
	Space     *SpaceHandler
	Doc       *DocumentHandler
	Version   *VersionHandler
	Search    *SearchHandler
	Workbench *WorkbenchHandler
	Upload    *UploadHandler
	Hub       *hub.Hub
	AuthSvc   *service.AuthService
	UploadDir string
}

func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Recover(), middleware.AccessLog(), middleware.CORS())

	r.GET("/api/v1/health", Health)
	r.Static("/uploads", d.UploadDir)

	api := r.Group("/api/v1")
	api.POST("/auth/register", d.Auth.Register)
	api.POST("/auth/login", d.Auth.Login)
	api.POST("/auth/refresh", d.Auth.Refresh)

	authed := api.Group("")
	authed.Use(middleware.Auth(d.AuthSvc))
	authed.GET("/auth/me", d.Auth.Me)
	authed.GET("/spaces", d.Space.List)
	authed.POST("/spaces", d.Space.Create)
	authed.GET("/spaces/:id", d.Space.Get)
	authed.PATCH("/spaces/:id", d.Space.Rename)
	authed.GET("/documents", d.Doc.Tree)
	authed.POST("/documents", d.Doc.Create)
	authed.GET("/documents/recycle", d.Doc.Recycle)
	authed.GET("/documents/:id", d.Doc.Get)
	authed.PATCH("/documents/:id", d.Doc.Update)
	authed.DELETE("/documents/:id", d.Doc.Delete)
	authed.POST("/documents/:id/move", d.Doc.Move)
	authed.POST("/documents/:id/restore", d.Doc.Restore)
	authed.POST("/documents/:id/favorite", d.Workbench.ToggleFavorite)
	authed.GET("/documents/:id/versions", d.Version.List)
	authed.POST("/documents/:id/versions", d.Version.Save)
	authed.GET("/versions/diff", d.Version.Diff)
	authed.POST("/versions/:id/rollback", d.Version.Rollback)
	authed.GET("/search", d.Search.Query)
	authed.GET("/workbench", d.Workbench.Home)
	authed.POST("/uploads", d.Upload.Image)

	r.GET("/ws", d.Hub.HandleWS)
	return r
}
