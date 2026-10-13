package service

import (
	"strings"
	"sync"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
)

// SessionStore 会话存储
type SessionStore struct {
	mu         sync.RWMutex
	sessions   map[string]*model.Session
	tasks      *TaskQueue
}

var store = &SessionStore{
	sessions: make(map[string]*model.Session),
	tasks:    NewTaskQueue(),
}

func GetSessionStore() *SessionStore {
	return store
}

func (ss *SessionStore) CreateSession(sessionID, title, bidNo string) *model.Session {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	session := &model.Session{
		SessionID: sessionID,
		Title:     title,
		BidNo:     bidNo,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	ss.sessions[sessionID] = session
	return session
}

func (ss *SessionStore) GetSession(sessionID string) (*model.Session, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	session, ok := ss.sessions[sessionID]
	return session, ok
}

func (ss *SessionStore) UpdateSession(session *model.Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	session.UpdatedAt = time.Now().UTC()
	ss.sessions[session.SessionID] = session
}

func (ss *SessionStore) SetParsingOutput(sessionID string, output *model.BidParsingOutput) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if session, ok := ss.sessions[sessionID]; ok {
		session.ParsingOutput = output
		session.UpdatedAt = time.Now().UTC()
	}
}

func (ss *SessionStore) SetOutline(sessionID string, outline *model.BidOutline) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if session, ok := ss.sessions[sessionID]; ok {
		session.Outline = outline
		session.Markdown = buildMarkdownFromOutline(outline)
		session.UpdatedAt = time.Now().UTC()
	}
}

func (ss *SessionStore) UpdateSectionContent(sessionID, sectionID, content string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if session, ok := ss.sessions[sessionID]; ok {
		if session.Markdown == "" {
			session.Markdown = "# " + sectionID + "\n\n" + content
		} else {
			session.Markdown += "\n\n" + content
		}
		session.CurrentSection = sectionID
		// 回填大纲章节正文,供按纲拼装导出
		if session.Outline != nil {
			for _, s := range session.Outline.Sections {
				if s.SectionID == sectionID {
					s.Content = content
					s.Status = model.ChapterStatusGenerated
					now := time.Now().UTC()
					s.GeneratedAt = &now
					break
				}
			}
		}
		session.UpdatedAt = time.Now().UTC()
	}
}

func (ss *SessionStore) SetSectionContent(sessionID, sectionID, title string, level int, content string, generatedAt *time.Time) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if session, ok := ss.sessions[sessionID]; ok {
		if session.Markdown == "" {
			indent := strings.Repeat("#", level)
			session.Markdown = indent + " " + title + "\n\n" + content
		} else {
			session.Markdown += "\n\n" + content
		}
		if session.Outline != nil {
			for _, s := range session.Outline.Sections {
				if s.SectionID == sectionID {
					s.Content = content
					s.Status = model.ChapterStatusGenerated
					s.GeneratedAt = generatedAt
					break
				}
			}
		}
		session.UpdatedAt = time.Now().UTC()
	}
}

func (ss *SessionStore) MarkSectionFailed(sessionID, sectionID, errMsg string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if session, ok := ss.sessions[sessionID]; ok {
		if session.Outline != nil {
			for _, s := range session.Outline.Sections {
				if s.SectionID == sectionID {
					s.Status = model.ChapterStatusFailed
					s.ResponseBasis = errMsg
					break
				}
			}
		}
		session.UpdatedAt = time.Now().UTC()
	}
}

func buildMarkdownFromOutline(outline *model.BidOutline) string {
	var md strings.Builder
	for _, section := range outline.Sections {
		indent := strings.Repeat("#", section.Level)
		md.WriteString(indent + " " + section.Title + "\n\n")
	}
	return md.String()
}
