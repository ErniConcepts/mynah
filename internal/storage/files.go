package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ErniConcepts/mynah/internal/llm"
)

type AgentPaths struct {
	RootPath           string
	UsersPath          string
	MemoryPath         string
	MemoryMetaPath     string
	MemoryRejectedPath string
	ProfilePath        string
	HistoryPath        string
}

type FileStore struct {
	paths            AgentPaths
	memoryCharLimit  int
	profileCharLimit int
	writeFile        func(path, content string) error
}

type RevisionProvenance struct {
	Target    string    `json:"target"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
}

type RejectedRevision struct {
	Timestamp      time.Time             `json:"timestamp"`
	UserID         string                `json:"user_id"`
	SessionID      string                `json:"session_id"`
	Message        string                `json:"message"`
	Reason         string                `json:"reason"`
	RejectionError string                `json:"rejection_error"`
	Operations     []llm.MemoryOperation `json:"operations"`
}

type AcceptedMemoryCommit struct {
	MemoryDoc        *string
	MemoryProvenance *RevisionProvenance
	UserDoc          *string
	UserProvenance   *RevisionProvenance
}

type RejectedMemoryCommit struct {
	RejectedRevision RejectedRevision
}

func NewAgentPaths(dataDir, tenantID, agentID string) AgentPaths {
	root := filepath.Join(dataDir, "tenants", tenantID, "agents", agentID)
	return AgentPaths{
		RootPath:           root,
		UsersPath:          filepath.Join(root, "users"),
		MemoryPath:         filepath.Join(root, "MEMORY.md"),
		MemoryMetaPath:     filepath.Join(root, "MEMORY.meta.json"),
		MemoryRejectedPath: filepath.Join(root, "MEMORY.rejected.json"),
		ProfilePath:        filepath.Join(root, "AGENT_PROFILE.md"),
		HistoryPath:        filepath.Join(root, "history.db"),
	}
}

func EnsureAgentPaths(paths AgentPaths) error {
	if err := os.MkdirAll(paths.RootPath, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.UsersPath, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(paths.MemoryPath); os.IsNotExist(err) {
		if err := os.WriteFile(paths.MemoryPath, []byte(""), 0o644); err != nil {
			return err
		}
	}
	if _, err := os.Stat(paths.ProfilePath); os.IsNotExist(err) {
		if err := os.WriteFile(paths.ProfilePath, []byte(""), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func NewFileStore(paths AgentPaths, memoryCharLimit, profileCharLimit int) *FileStore {
	return &FileStore{
		paths:            paths,
		memoryCharLimit:  memoryCharLimit,
		profileCharLimit: profileCharLimit,
		writeFile:        writeAtomically,
	}
}

func (s *FileStore) ReadMemory() (string, error) {
	return readTrimmed(s.paths.MemoryPath)
}

func (s *FileStore) ReadProfile() (string, error) {
	return readTrimmed(s.paths.ProfilePath)
}

func (s *FileStore) ReadUserProfile(userID string) (string, error) {
	return readTrimmed(s.userProfilePath(userID))
}

func (s *FileStore) WriteMemory(content string) error {
	return s.writeFile(s.paths.MemoryPath, strings.TrimSpace(content)+"\n")
}

func (s *FileStore) WriteProfile(content string) error {
	return s.writeFile(s.paths.ProfilePath, strings.TrimSpace(content)+"\n")
}

func (s *FileStore) ReadMemoryProvenance() (RevisionProvenance, error) {
	return readProvenance(s.paths.MemoryMetaPath)
}

func (s *FileStore) WriteMemoryProvenance(meta RevisionProvenance) error {
	meta.Target = "memory"
	return s.writeJSON(s.paths.MemoryMetaPath, meta)
}

func (s *FileStore) ReadRejectedRevision() (RejectedRevision, error) {
	return readJSON[RejectedRevision](s.paths.MemoryRejectedPath)
}

func (s *FileStore) WriteRejectedRevision(rejected RejectedRevision) error {
	return s.writeJSON(s.paths.MemoryRejectedPath, rejected)
}

func (s *FileStore) WriteUserProfile(userID, content string) error {
	path := s.userProfilePath(userID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return s.writeFile(path, strings.TrimSpace(content)+"\n")
}

func (s *FileStore) ReadUserProfileProvenance(userID string) (RevisionProvenance, error) {
	return readProvenance(s.userProfileMetaPath(userID))
}

func (s *FileStore) WriteUserProfileProvenance(userID string, meta RevisionProvenance) error {
	path := s.userProfileMetaPath(userID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	meta.Target = "user"
	meta.UserID = strings.TrimSpace(userID)
	return s.writeJSON(path, meta)
}

func (s *FileStore) CommitAcceptedMemory(userID string, commit AcceptedMemoryCommit) error {
	type writeOp struct {
		path    string
		content string
		exists  bool
	}
	ops := make([]writeOp, 0, 4)

	if commit.MemoryDoc != nil {
		ops = append(ops, writeOp{
			path:    s.paths.MemoryPath,
			content: strings.TrimSpace(*commit.MemoryDoc) + "\n",
		})
	}
	if commit.MemoryProvenance != nil {
		meta := *commit.MemoryProvenance
		meta.Target = "memory"
		payload, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			return err
		}
		ops = append(ops, writeOp{
			path:    s.paths.MemoryMetaPath,
			content: string(payload) + "\n",
		})
	}
	if commit.UserDoc != nil {
		path := s.userProfilePath(userID)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		ops = append(ops, writeOp{
			path:    path,
			content: strings.TrimSpace(*commit.UserDoc) + "\n",
		})
	}
	if commit.UserProvenance != nil {
		path := s.userProfileMetaPath(userID)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		meta := *commit.UserProvenance
		meta.Target = "user"
		meta.UserID = strings.TrimSpace(userID)
		payload, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			return err
		}
		ops = append(ops, writeOp{
			path:    path,
			content: string(payload) + "\n",
		})
	}
	if len(ops) == 0 {
		return nil
	}

	originals := make([]writeOp, 0, len(ops))
	applied := make([]writeOp, 0, len(ops))
	for _, op := range ops {
		original, err := os.ReadFile(op.path)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			original = nil
		}
		originals = append(originals, writeOp{path: op.path, content: string(original), exists: err == nil})

		if err := s.writeFile(op.path, op.content); err != nil {
			for index := len(applied) - 1; index >= 0; index-- {
				restore := originals[index]
				var rollbackErr error
				if !restore.exists {
					rollbackErr = os.Remove(restore.path)
					if rollbackErr != nil && !os.IsNotExist(rollbackErr) {
						return fmt.Errorf("commit accepted memory: %w; rollback failed for %s: %v", err, restore.path, rollbackErr)
					}
					continue
				}
				if rollbackErr = s.writeFile(restore.path, restore.content); rollbackErr != nil {
					return fmt.Errorf("commit accepted memory: %w; rollback failed for %s: %v", err, restore.path, rollbackErr)
				}
			}
			return fmt.Errorf("commit accepted memory: %w", err)
		}
		applied = append(applied, op)
	}
	return nil
}

func (s *FileStore) CommitRejectedMemory(commit RejectedMemoryCommit) error {
	return s.writeJSON(s.paths.MemoryRejectedPath, commit.RejectedRevision)
}

func (s *FileStore) writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return s.writeFile(path, string(payload)+"\n")
}

func readTrimmed(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func (s *FileStore) userProfilePath(userID string) string {
	return filepath.Join(s.paths.UsersPath, userID, "USER.md")
}

func (s *FileStore) userProfileMetaPath(userID string) string {
	return filepath.Join(s.paths.UsersPath, userID, "USER.meta.json")
}

func writeAtomically(path, content string) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".mynah-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()

	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	if _, err := temp.WriteString(content); err != nil {
		_ = temp.Close()
		cleanup()
		return err
	}
	if err := temp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

func writeJSONAtomically(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomically(path, string(payload)+"\n")
}

func readProvenance(path string) (RevisionProvenance, error) {
	return readJSON[RevisionProvenance](path)
}

func readJSON[T any](path string) (T, error) {
	var value T
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return value, nil
		}
		return value, err
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, err
	}
	return value, nil
}
