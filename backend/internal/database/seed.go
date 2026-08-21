package database

import (
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gowiki/internal/logger"
	"gowiki/internal/model"
	"gowiki/internal/pkg/timeutil"
)

const (
	SeedAdminEmail  = "admin@gowiki.dev"
	SeedAdminPass   = "admin123"
	SeedEditorEmail = "editor@gowiki.dev"
	SeedEditorPass  = "editor123"
)

func Seed(db *gorm.DB) error {
	var n int64
	if err := db.Model(&model.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	adminHash, _ := bcrypt.GenerateFromPassword([]byte(SeedAdminPass), bcrypt.DefaultCost)
	editorHash, _ := bcrypt.GenerateFromPassword([]byte(SeedEditorPass), bcrypt.DefaultCost)
	admin := model.User{
		Email: SeedAdminEmail, PasswordHash: string(adminHash),
		DisplayName: "林昭", AvatarColor: "#C45C26",
	}
	editor := model.User{
		Email: SeedEditorEmail, PasswordHash: string(editorHash),
		DisplayName: "沈清", AvatarColor: "#2F6F4E",
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	if err := db.Create(&editor).Error; err != nil {
		return err
	}
	space := model.Space{Name: "产品知识库", OwnerID: admin.ID}
	if err := db.Create(&space).Error; err != nil {
		return err
	}
	welcomeID := uuid.New()
	guideID := uuid.New()
	noteID := uuid.New()
	welcome := model.Document{
		ID: welcomeID, SpaceID: space.ID, Title: "欢迎使用 GoWiki",
		Path: "/" + welcomeID.String() + "/", SortOrder: 0,
		EditorMode: model.ModeMarkdown,
		ContentMD:  welcomeMD,
		ContentJSON: "",
	}
	guide := model.Document{
		ID: guideID, SpaceID: space.ID, Title: "工程规范",
		Path: "/" + guideID.String() + "/", SortOrder: 1,
		EditorMode: model.ModeMarkdown,
		ContentMD:  guideMD,
	}
	parent := welcomeID
	note := model.Document{
		ID: noteID, SpaceID: space.ID, ParentID: &parent, Title: "第一次编辑会议",
		Path: welcome.Path + noteID.String() + "/", SortOrder: 0,
		EditorMode: model.ModeMarkdown,
		ContentMD:  noteMD,
	}
	if err := db.Create(&[]model.Document{welcome, guide, note}).Error; err != nil {
		return err
	}
	now := timeutil.Now()
	_ = db.Create(&model.Activity{
		SpaceID: space.ID, ActorID: admin.ID, Action: "seed",
		DocumentID: &welcomeID, Summary: "初始化产品知识库", CreatedAt: now,
	}).Error
	logger.L().Info("seeded demo workspace", "admin", SeedAdminEmail, "editor", SeedEditorEmail)
	return nil
}

var welcomeMD = strings.TrimSpace(`
# 欢迎使用 GoWiki

这是一份可协同编辑的团队知识库。左侧是**无限层级文档树**，中间是 Markdown / 富文本编辑器。

## 快速开始

1. 在树节点上新建子文档
2. 拖拽调整层级（不能拖进自己的子树）
3. Markdown 模式下多人可同时输入，变更会实时合并
4. 富文本模式按段落加锁，避免光标打架
5. 需要里程碑时点「保存版本」，之后可 Diff 与回滚

试试搜索「协同」或「版本」。
`)

var guideMD = strings.TrimSpace(`
# 工程规范

- 业务时间一律使用北京时间（GMT+8）
- 后端禁止 CGO 依赖，保证 ARM64 / AMD64 交叉编译
- 版本回滚采用 Git revert 语义：以历史内容创建新版本
`)

var noteMD = strings.TrimSpace(`
# 第一次编辑会议

- 确认协同走自研 RGA，不做 Yjs 绑定
- 检索使用 Bleve + gse
- 工作台展示最近浏览、收藏与动态
`)
