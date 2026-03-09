package docker

import "encoding/json"

type SMTPSettings struct {
	Server   string `json:"s,omitempty"`
	Port     string `json:"p,omitempty"`
	Username string `json:"u,omitempty"`
	Password string `json:"pw,omitempty"`
	From     string `json:"f,omitempty"`
}

func (s SMTPSettings) BuildEnv() []string {
	if s.Server == "" {
		return nil
	}
	return []string{
		"SMTP_ADDRESS=" + s.Server,
		"SMTP_PORT=" + s.Port,
		"SMTP_USERNAME=" + s.Username,
		"SMTP_PASSWORD=" + s.Password,
		"MAILER_FROM_ADDRESS=" + s.From,
	}
}

type ContainerResources struct {
	CPUs     int `json:"cpus,omitempty"`
	MemoryMB int `json:"mem,omitempty"`
}

type BackupSettings struct {
	Path     string `json:"p,omitempty"`
	AutoBack bool   `json:"a,omitempty"`
}

type ApplicationSettings struct {
	Name       string             `json:"n"`
	Image      string             `json:"i"`
	Host       string             `json:"h"`
	DisableTLS bool               `json:"dt"`
	EnvVars    map[string]string  `json:"env"`
	SMTP       SMTPSettings       `json:"sm"`
	Resources  ContainerResources `json:"res"`
	AutoUpdate bool               `json:"au"`
	Backup     BackupSettings     `json:"bk"`
}

func UnmarshalApplicationSettings(s string) (ApplicationSettings, error) {
	var settings ApplicationSettings
	err := json.Unmarshal([]byte(s), &settings)
	return settings, err
}

func (s ApplicationSettings) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s ApplicationSettings) TLSEnabled() bool {
	return s.Host != "" && !s.DisableTLS && !IsLocalhost(s.Host)
}

func (s ApplicationSettings) Equal(other ApplicationSettings) bool {
	if s.Name != other.Name || s.Image != other.Image || s.Host != other.Host || s.DisableTLS != other.DisableTLS {
		return false
	}
	if s.Resources != other.Resources {
		return false
	}
	if s.SMTP != other.SMTP {
		return false
	}
	if s.AutoUpdate != other.AutoUpdate {
		return false
	}
	if s.Backup != other.Backup {
		return false
	}
	if len(s.EnvVars) != len(other.EnvVars) {
		return false
	}
	for k, v := range s.EnvVars {
		if other.EnvVars[k] != v {
			return false
		}
	}
	return true
}

func (s ApplicationSettings) BuildEnv(secretKeyBase string) []string {
	env := []string{
		"SECRET_KEY_BASE=" + secretKeyBase,
	}

	if !s.TLSEnabled() {
		env = append(env, "DISABLE_SSL=true")
	}

	env = append(env, s.SMTP.BuildEnv()...)

	for k, v := range s.EnvVars {
		env = append(env, k+"="+v)
	}

	return env
}
