package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalSkill represents a skill found in the system's skills directory
type LocalSkill struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SkillService handles scanning and managing local skills
type SkillService struct {
	skillsDir string
}

// NewSkillService creates a new skill service with the given skills directory
func NewSkillService(skillsDir string) *SkillService {
	return &SkillService{
		skillsDir: skillsDir,
	}
}

// ScanLocalSkills scans the skills directory for SKILL.md files and parses their frontmatter
func (s *SkillService) ScanLocalSkills() ([]LocalSkill, error) {
	var skills []LocalSkill

	entries, err := os.ReadDir(s.skillsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillPath := filepath.Join(s.skillsDir, entry.Name(), "SKILL.md")
			content, err := os.ReadFile(skillPath)
			if err != nil {
				continue // Skip unreadable files
			}

			skill := s.parseFrontmatter(entry.Name(), string(content))
			if skill.Name != "" {
				skills = append(skills, skill)
			}
		}
	}

	return skills, nil
}

// parseFrontmatter extracts name and description from YAML frontmatter
func (s *SkillService) parseFrontmatter(name, content string) LocalSkill {
	skill := LocalSkill{
		Path: name,
		Name: name,
	}

	// Check for YAML frontmatter (--- delimited)
	if !strings.HasPrefix(content, "---") {
		return skill
	}

	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return skill
	}

	fmContent := content[3 : endIdx+3]
	lines := strings.Split(fmContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])

		// Remove quotes if present
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")

		switch key {
		case "name":
			if value != "" {
				skill.Name = value
			}
		case "description":
			skill.Description = value
		}
	}

	return skill
}

// GetSkillContent returns the full content of a skill's SKILL.md file
func (s *SkillService) GetSkillContent(skillName string) (string, error) {
	skillPath := filepath.Join(s.skillsDir, skillName, "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("failed to read skill file: %v", err)
	}
	return string(content), nil
}
