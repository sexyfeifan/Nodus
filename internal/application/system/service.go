package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nodus/pkg/buildinfo"

	"github.com/pocketbase/pocketbase/core"
)

const SettingsID = "wcstsqmz8hur331"

type Service struct {
	app core.App
}

type InitializeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language"`
}

type SettingsResponse struct {
	ID          string                 `json:"id"`
	Initialized bool                   `json:"initialized"`
	General     map[string]interface{} `json:"general"`
}

type UpdateSettingsRequest struct {
	General map[string]interface{} `json:"general"`
}

func NewService(app core.App) *Service {
	return &Service{app: app}
}

// checkInitialized checks if the system has been initialized against the given app (or transaction).
func checkInitialized(app core.App) (bool, error) {
	collection, err := app.FindCollectionByNameOrId("fh_settings")
	if err != nil {
		return false, fmt.Errorf("failed to find settings collection: %w", err)
	}

	record, err := app.FindRecordById(collection, SettingsID)
	if err != nil {
		// If record doesn't exist, system is not initialized
		return false, nil
	}

	initialized := record.GetBool("initialized")
	return initialized, nil
}

// CheckInitialized checks if the system has been initialized.
func (s *Service) CheckInitialized() (bool, error) {
	return checkInitialized(s.app)
}

// GetSettings retrieves the system settings.
func (s *Service) GetSettings() (*SettingsResponse, error) {
	collection, err := s.app.FindCollectionByNameOrId("fh_settings")
	if err != nil {
		return nil, fmt.Errorf("failed to find settings collection: %w", err)
	}

	record, err := s.app.FindRecordById(collection, SettingsID)
	if err != nil {
		return nil, fmt.Errorf("failed to find settings record: %w", err)
	}

	general := make(map[string]interface{})
	if generalField := record.Get("general"); generalField != nil {
		if generalMap, ok := generalField.(map[string]interface{}); ok {
			general = generalMap
		} else if generalStr, ok := generalField.(string); ok {
			if err := json.Unmarshal([]byte(generalStr), &general); err != nil {
				s.app.Logger().Warn("Failed to unmarshal general settings", "error", err)
			}
		}
	}

	return &SettingsResponse{
		ID:          record.Id,
		Initialized: record.GetBool("initialized"),
		General:     general,
	}, nil
}

// UpdateSettings updates the system settings.
func (s *Service) UpdateSettings(req *UpdateSettingsRequest) error {
	if req.General == nil {
		return errors.New("general settings cannot be nil")
	}

	generalJSON, err := json.Marshal(req.General)
	if err != nil {
		return fmt.Errorf("failed to marshal general settings: %w", err)
	}

	_, err = s.app.DB().NewQuery(`
		UPDATE fh_settings
		SET general = {:general}, updated = STRFTIME('%Y-%m-%d %H:%M:%S', 'now')
		WHERE id = {:id}
	`).Bind(map[string]interface{}{
		"general": string(generalJSON),
		"id":      SettingsID,
	}).Execute()

	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	s.app.Logger().Info("Settings updated successfully")
	return nil
}

// GetMonitoringIntervals reads latencyCheckInterval and locationCheckInterval
// from settings general JSON. Falls back to defaults if not set or on error.
func (s *Service) GetMonitoringIntervals() (latencyInterval, geoInterval time.Duration) {
	const defaultLatency = 5 * time.Second
	const defaultGeo = 24 * time.Hour

	settings, err := s.GetSettings()
	if err != nil {
		s.app.Logger().Warn("Failed to load settings for monitoring intervals, using defaults", "error", err)
		return defaultLatency, defaultGeo
	}

	latencyInterval = defaultLatency
	if v, ok := settings.General["latencyCheckInterval"]; ok {
		if secs, ok := toFloat64(v); ok && secs >= 1 && secs <= 24*3600 {
			latencyInterval = time.Duration(secs) * time.Second
		}
	}

	geoInterval = defaultGeo
	if v, ok := settings.General["locationCheckInterval"]; ok {
		if secs, ok := toFloat64(v); ok && secs >= 1 && secs <= 30*24*3600 {
			geoInterval = time.Duration(secs) * time.Second
		}
	}

	return latencyInterval, geoInterval
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// GetAppVersion returns the application version.
func (s *Service) GetAppVersion() string {
	return buildinfo.AppVersion
}

// LatestVersionResponse holds the latest release info from GitHub.
type LatestVersionResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// GetLatestVersion fetches the latest Nodus release from GitHub.
func (s *Service) GetLatestVersion() (*LatestVersionResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/sexyfeifan/Nodus/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Nodus-server/"+buildinfo.AppVersion)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var release LatestVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// initializeMu serializes concurrent initialization attempts to prevent
// duplicate admin accounts racing past the initialized check.
var initializeMu sync.Mutex

// InitializeSystem initializes the system with admin user and settings (with transaction).
func (s *Service) InitializeSystem(req *InitializeRequest) error {
	if req.Email == "" || req.Password == "" {
		return errors.New("email and password are required")
	}

	if req.Language != "en" && req.Language != "zh" {
		req.Language = "en"
	}

	initializeMu.Lock()
	defer initializeMu.Unlock()

	return s.app.RunInTransaction(func(txApp core.App) error {
		initialized, err := checkInitialized(txApp)
		if err != nil {
			return fmt.Errorf("failed to check initialization status: %w", err)
		}
		if initialized {
			return errors.New("system is already initialized")
		}

		usersCollection, err := txApp.FindCollectionByNameOrId("fh_users")
		if err != nil {
			return fmt.Errorf("failed to find users collection: %w", err)
		}

		userRecord := core.NewRecord(usersCollection)
		userRecord.Set("email", req.Email)
		userRecord.Set("password", req.Password)
		userRecord.Set("passwordConfirm", req.Password)
		userRecord.Set("emailVisibility", true)
		userRecord.Set("role", "admin")

		if err := txApp.Save(userRecord); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		settingsCollection, err := txApp.FindCollectionByNameOrId("fh_settings")
		if err != nil {
			return fmt.Errorf("failed to find settings collection: %w", err)
		}

		settingsRecord, err := txApp.FindRecordById(settingsCollection, SettingsID)
		if err != nil {
			settingsRecord = core.NewRecord(settingsCollection)
			settingsRecord.Set("id", SettingsID)
		}

		settingsRecord.Set("initialized", true)
		settingsRecord.Set("general", map[string]interface{}{
			"defaultLanguage": req.Language,
		})

		if err := txApp.Save(settingsRecord); err != nil {
			return fmt.Errorf("failed to save settings: %w", err)
		}

		s.app.Logger().Info("System initialized successfully",
			"email", req.Email,
			"language", req.Language,
		)

		return nil
	})
}
