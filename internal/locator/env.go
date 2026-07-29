package locator

// SessionLocator 负责定位会话文件
import (
	"os"
	"path/filepath"
)

const projectsRootEnv = "META_CC_PROJECTS_ROOT"
const codexHomeEnv = "CODEX_HOME"

const (
	HostClaudeCode = "claude"
	HostCodex      = "codex"
	HostOverride   = "override"
)

type SessionRoot struct {
	Host          string
	Path          string
	ProjectHashed bool
}

type SessionLocator struct {
	projectsRoot string
	roots        []SessionRoot
}

// NewSessionLocator 创建 SessionLocator 实例
func NewSessionLocator() *SessionLocator {
	roots := DefaultSessionRoots()
	root := os.Getenv(projectsRootEnv)
	if root == "" {
		for _, candidate := range roots {
			if candidate.Host == HostClaudeCode {
				root = candidate.Path
				break
			}
		}
	} else {
		root = filepath.Clean(root)
		roots = append([]SessionRoot{{
			Host:          HostOverride,
			Path:          root,
			ProjectHashed: true,
		}}, roots...)
	}

	return &SessionLocator{
		projectsRoot: root,
		roots:        roots,
	}
}

func DefaultSessionRoots() []SessionRoot {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil
		}
	}

	codexHome := resolveCodexRoot()

	return []SessionRoot{
		{
			Host:          HostClaudeCode,
			Path:          filepath.Join(homeDir, ".claude", "projects"),
			ProjectHashed: true,
		},
		{
			Host:          HostCodex,
			Path:          filepath.Join(codexHome, "sessions"),
			ProjectHashed: false,
		},
	}
}

func (l *SessionLocator) TranscriptRoots() []SessionRoot {
	roots := make([]SessionRoot, len(l.roots))
	copy(roots, l.roots)
	return roots
}
