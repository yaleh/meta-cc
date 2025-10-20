package locator

// SessionLocator 负责定位会话文件
type SessionLocator struct{}

// NewSessionLocator 创建 SessionLocator 实例
func NewSessionLocator() *SessionLocator {
	return &SessionLocator{}
}
