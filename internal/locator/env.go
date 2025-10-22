package locator

// SessionLocator 负责定位会话文件
import (
	"os"
	"path/filepath"
)

const projectsRootEnv = "META_CC_PROJECTS_ROOT"

type SessionLocator struct {
	projectsRoot string
}

// NewSessionLocator 创建 SessionLocator 实例
func NewSessionLocator() *SessionLocator {
	root := os.Getenv(projectsRootEnv)
	if root == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			root = filepath.Join(homeDir, ".claude", "projects")
		}
	} else {
		root = filepath.Clean(root)
	}

	return &SessionLocator{
		projectsRoot: root,
	}
}
