package skillinsights

import "time"

type Options struct {
	Roots       []string
	SkillRoots  []string
	KnownSkills []string
	Since       time.Time
	MaxFiles    int
	MaxFileMB   int64
	MaxSamples  int
}

type Report struct {
	GeneratedAt       time.Time           `json:"generated_at"`
	Since             time.Time           `json:"since"`
	FilesScanned      int                 `json:"files_scanned"`
	FilesSkipped      int                 `json:"files_skipped"`
	FilesWithErrors   int                 `json:"files_with_errors"`
	SessionsScanned   int                 `json:"sessions_scanned"`
	SkillUses         []SkillUse          `json:"skill_uses"`
	MissedSkills      []MissedSkill       `json:"missed_skills"`
	PossibleMisuses   []PossibleMisuse    `json:"possible_misuses"`
	SkillStats        []SkillStat         `json:"skill_stats"`
	ScenarioStats     []ScenarioSkillStat `json:"scenario_stats"`
	ToolErrorStats    []ToolErrorStat     `json:"tool_error_stats"`
	UnreadableSamples []FileError         `json:"unreadable_samples,omitempty"`
}

type SkillUse struct {
	Skill       string    `json:"skill"`
	Scenario    string    `json:"scenario"`
	TriggerType string    `json:"trigger_type"`
	Confidence  float64   `json:"confidence"`
	Effective   string    `json:"effective"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"session_id"`
	Provider    string    `json:"provider"`
	File        string    `json:"file"`
	Evidence    string    `json:"evidence"`
	ErrorCount  int       `json:"error_count"`
}

type MissedSkill struct {
	ExpectedSkill string    `json:"expected_skill"`
	Scenario      string    `json:"scenario"`
	Reason        string    `json:"reason"`
	Confidence    float64   `json:"confidence"`
	Timestamp     time.Time `json:"timestamp"`
	SessionID     string    `json:"session_id"`
	Provider      string    `json:"provider"`
	File          string    `json:"file"`
	Evidence      string    `json:"evidence"`
	ErrorCount    int       `json:"error_count"`
}

type PossibleMisuse struct {
	Skill      string    `json:"skill"`
	Scenario   string    `json:"scenario"`
	Reason     string    `json:"reason"`
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
	SessionID  string    `json:"session_id"`
	Provider   string    `json:"provider"`
	File       string    `json:"file"`
	Evidence   string    `json:"evidence"`
}

type SkillStat struct {
	Skill        string  `json:"skill"`
	Uses         int     `json:"uses"`
	Effective    int     `json:"effective"`
	Ineffective  int     `json:"ineffective"`
	Unknown      int     `json:"unknown"`
	EffectRate   float64 `json:"effect_rate"`
	TopScenarios []Count `json:"top_scenarios"`
}

type ScenarioSkillStat struct {
	Scenario string `json:"scenario"`
	Skill    string `json:"skill"`
	Uses     int    `json:"uses"`
}

type ToolErrorStat struct {
	Tool   string `json:"tool"`
	Errors int    `json:"errors"`
}

type Count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type FileError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}
