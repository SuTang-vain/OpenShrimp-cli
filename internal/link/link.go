package link

import (
	"fmt"
	"os"

	"ai-manager/internal/config"
	"ai-manager/internal/utils"
)

type LinkManager struct {
	cfg *config.Config
}

func NewLinkManager(cfg *config.Config) *LinkManager {
	return &LinkManager{cfg: cfg}
}

type LinkInfo struct {
	ToolName   string `json:"tool_name"`
	LinkPath   string `json:"link_path"`
	TargetPath string `json:"target_path"`
	Exists     bool   `json:"exists"`
	Valid      bool   `json:"valid"`
	IsSymlink  bool   `json:"is_symlink"`
	Error      string `json:"error,omitempty"`
}

type LinkResult struct {
	Tool     string `json:"tool"`
	LinkPath string `json:"link_path"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

func (m *LinkManager) ListLinks() ([]LinkInfo, error) {
	links := make([]LinkInfo, 0)

	for toolName, tool := range m.cfg.Tools {
		linkPath := m.getLinkPath(toolName)
		targetPath := m.getTargetPath(toolName)

		info := LinkInfo{
			ToolName:   tool.Name,
			LinkPath:   linkPath,
			TargetPath: targetPath,
		}

		expandedLink := utils.ExpandPath(linkPath)
		expandedTarget := utils.ExpandPath(targetPath)

		linkStat, err := os.Lstat(expandedLink)
		if err != nil {
			info.Exists = false
			info.Valid = false
		} else {
			info.Exists = true
			info.IsSymlink = linkStat.Mode()&os.ModeSymlink != 0

			if info.IsSymlink {
				target, err := os.Readlink(expandedLink)
				if err != nil {
					info.Valid = false
					info.Error = err.Error()
				} else {
					info.Valid = (target == targetPath) || (target == expandedTarget)
					if !info.Valid {
						info.Error = fmt.Sprintf("target mismatch: %s", target)
					}
				}
			} else {
				info.Valid = false
				info.Error = "not a symbolic link"
			}
		}

		links = append(links, info)
	}

	return links, nil
}

func (m *LinkManager) CreateLink(toolName string) (*LinkResult, error) {
	tool, ok := m.cfg.Tools[toolName]
	if !ok {
		return &LinkResult{
			Tool:    toolName,
			Success: false,
			Message: fmt.Sprintf("tool not found: %s", toolName),
		}, fmt.Errorf("tool not found: %s", toolName)
	}

	linkPath := utils.ExpandPath(m.getLinkPath(toolName))
	targetPath := utils.ExpandPath(m.getTargetPath(toolName))

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("target path does not exist: %s", targetPath),
		}, fmt.Errorf("target path does not exist: %s", targetPath)
	}

	if err := utils.EnsureDir(linkPath); err != nil {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("failed to create directory: %v", err),
		}, err
	}

	if err := utils.CreateSymlink(targetPath, linkPath); err != nil {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("failed to create symlink: %v", err),
		}, err
	}

	return &LinkResult{
		Tool:     tool.Name,
		LinkPath: linkPath,
		Success:  true,
		Message:  fmt.Sprintf("created symlink: %s -> %s", linkPath, targetPath),
	}, nil
}

func (m *LinkManager) RemoveLink(toolName string) (*LinkResult, error) {
	tool, ok := m.cfg.Tools[toolName]
	if !ok {
		return &LinkResult{
			Tool:    toolName,
			Success: false,
			Message: fmt.Sprintf("tool not found: %s", toolName),
		}, fmt.Errorf("tool not found: %s", toolName)
	}

	linkPath := utils.ExpandPath(m.getLinkPath(toolName))

	if _, err := os.Lstat(linkPath); os.IsNotExist(err) {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  true,
			Message:  "link does not exist, nothing to remove",
		}, nil
	}

	if !utils.IsSymlink(linkPath) {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  "path exists but is not a symlink",
		}, fmt.Errorf("path exists but is not a symlink: %s", linkPath)
	}

	if err := os.Remove(linkPath); err != nil {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("failed to remove link: %v", err),
		}, err
	}

	return &LinkResult{
		Tool:     tool.Name,
		LinkPath: linkPath,
		Success:  true,
		Message:  fmt.Sprintf("removed symlink: %s", linkPath),
	}, nil
}

func (m *LinkManager) VerifyLink(toolName string) (*LinkResult, error) {
	tool, ok := m.cfg.Tools[toolName]
	if !ok {
		return &LinkResult{
			Tool:    toolName,
			Success: false,
			Message: fmt.Sprintf("tool not found: %s", toolName),
		}, fmt.Errorf("tool not found: %s", toolName)
	}

	linkPath := utils.ExpandPath(m.getLinkPath(toolName))

	exists, target, err := utils.CheckSymlink(linkPath)
	if err != nil {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("link invalid: %v", err),
		}, nil
	}

	if !exists {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("link broken: target '%s' does not exist", target),
		}, nil
	}

	expectedTarget := utils.ExpandPath(m.getTargetPath(toolName))
	if target != expectedTarget && target != m.getTargetPath(toolName) {
		return &LinkResult{
			Tool:     tool.Name,
			LinkPath: linkPath,
			Success:  false,
			Message:  fmt.Sprintf("link points to wrong target: %s (expected: %s)", target, expectedTarget),
		}, nil
	}

	return &LinkResult{
		Tool:     tool.Name,
		LinkPath: linkPath,
		Success:  true,
		Message:  fmt.Sprintf("link valid: %s -> %s", linkPath, target),
	}, nil
}

func (m *LinkManager) CreateAllLinks() ([]LinkResult, error) {
	results := make([]LinkResult, 0)

	for toolName := range m.cfg.Tools {
		result, err := m.CreateLink(toolName)
		results = append(results, *result)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func (m *LinkManager) RemoveAllLinks() ([]LinkResult, error) {
	results := make([]LinkResult, 0)

	for toolName := range m.cfg.Tools {
		result, err := m.RemoveLink(toolName)
		results = append(results, *result)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func (m *LinkManager) getLinkPath(toolName string) string {
	return fmt.Sprintf("~/.ai-manager/links/%s", toolName)
}

func (m *LinkManager) getTargetPath(toolName string) string {
	tool, ok := m.cfg.Tools[toolName]
	if !ok {
		return ""
	}
	return tool.Path
}

func (m *LinkManager) InitializeLinksDir() error {
	linksDir := utils.ExpandPath("~/.ai-manager/links")
	if err := os.MkdirAll(linksDir, 0755); err != nil {
		return fmt.Errorf("failed to create links directory: %w", err)
	}
	return nil
}

type LinkStatus string

const (
	LinkStatusOK      LinkStatus = "ok"
	LinkStatusMissing LinkStatus = "missing"
	LinkStatusBroken  LinkStatus = "broken"
	LinkStatusWrong   LinkStatus = "wrong_target"
	LinkStatusNotLink LinkStatus = "not_symlink"
)

func (m *LinkManager) GetLinkStatus(toolName string) (LinkStatus, string, error) {
	tool, ok := m.cfg.Tools[toolName]
	if !ok {
		return LinkStatusMissing, "", fmt.Errorf("tool not found: %s", toolName)
	}

	linkPath := utils.ExpandPath(m.getLinkPath(toolName))

	linkStat, err := os.Lstat(linkPath)
	if os.IsNotExist(err) {
		return LinkStatusMissing, "link does not exist", nil
	}
	if err != nil {
		return LinkStatusMissing, "", err
	}

	if linkStat.Mode()&os.ModeSymlink == 0 {
		return LinkStatusNotLink, "path exists but is not a symlink", nil
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		return LinkStatusBroken, "failed to read link target", err
	}

	expectedTarget := utils.ExpandPath(tool.Path)
	if target != expectedTarget && target != tool.Path {
		return LinkStatusWrong, fmt.Sprintf("target mismatch: %s", target), nil
	}

	if _, err := os.Stat(expectedTarget); os.IsNotExist(err) {
		return LinkStatusBroken, "target does not exist", nil
	}

	return LinkStatusOK, "", nil
}
