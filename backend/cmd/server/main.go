package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gowiki/internal/collab/hub"
	"gowiki/internal/config"
	"gowiki/internal/database"
	"gowiki/internal/handler"
	"gowiki/internal/logger"
	"gowiki/internal/repository"
	"gowiki/internal/search"
	"gowiki/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.Init(cfg.LogLevel, cfg.Env)
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Error("mkdir upload", "err", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	if err := database.Seed(db); err != nil {
		log.Error("seed", "err", err)
		os.Exit(1)
	}

	eng, err := search.Open(cfg.BlevePath, cfg.SearchAnalyzer)
	if err != nil {
		log.Error("bleve", "err", err)
		os.Exit(1)
	}
	defer eng.Close()

	users := repository.NewUserRepo(db)
	spaces := repository.NewSpaceRepo(db)
	docs := repository.NewDocumentRepo(db)
	vers := repository.NewVersionRepo(db)
	ops := repository.NewOpRepo(db)
	acts := repository.NewActivityRepo(db)
	wb := repository.NewWorkbenchRepo(db)

	authSvc := service.NewAuthService(users, cfg)
	spaceSvc := service.NewSpaceService(spaces)
	docSvc := service.NewDocumentService(docs, spaces, acts, wb, eng)
	treeSvc := service.NewTreeService(docs)
	verSvc := service.NewVersionService(vers, docs, acts, eng)
	wbSvc := service.NewWorkbenchService(wb, docs, acts)
	upSvc := service.NewUploadService(cfg)

	if err := docSvc.ReindexAll(); err != nil {
		log.Warn("reindex", "err", err)
	}

	h := hub.New(docs, ops, verSvc, hubParser{authSvc}, time.Duration(cfg.LockTimeoutS)*time.Second)

	r := handler.NewRouter(handler.Deps{
		Auth:      handler.NewAuthHandler(authSvc),
		Space:     handler.NewSpaceHandler(spaceSvc),
		Doc:       handler.NewDocumentHandler(docSvc, treeSvc, wbSvc),
		Version:   handler.NewVersionHandler(verSvc),
		Search:    handler.NewSearchHandler(eng),
		Workbench: handler.NewWorkbenchHandler(wbSvc),
		Upload:    handler.NewUploadHandler(upSvc),
		Hub:       h,
		AuthSvc:   authSvc,
		UploadDir: cfg.UploadDir,
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadHeaderTimeout: 8 * time.Second}
	go func() {
		log.Info("gowiki listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http", "err", err)
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Close()
}

type hubParser struct{ auth *service.AuthService }

func (p hubParser) Parse(token string) (userID, name, color string, err error) {
	c, err := p.auth.Parse(token)
	if err != nil {
		return "", "", "", err
	}
	return c.UserID, c.DisplayName, "#C45C26", nil
}
