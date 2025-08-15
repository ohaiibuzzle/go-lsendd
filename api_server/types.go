package apiserver

import "time"

type RegisterMessage struct {
	Alias       string `json:"alias"`
	Version     string `json:"version"`
	DeviceModel string `json:"device_model"`
	DeviceType  string `json:"device_type"`
	Fingerprint string `json:"fingerprint"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Download    bool   `json:"download"`
}

type FileMetadata struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Size     int    `json:"size"`
	FileType string `json:"fileType"`
	Sha256   string `json:"sha256"`
	Preview  string `json:"preview"`
	Metadata struct {
		Modified time.Time `json:"modified"`
		Accessed time.Time `json:"accessed"`
	} `json:"metadata,omitempty"`
}

type UploadMetadataMessage struct {
	Info struct {
		Alias       string `json:"alias"`
		Version     string `json:"version"`
		DeviceModel string `json:"deviceModel"`
		DeviceType  string `json:"deviceType"`
		Fingerprint string `json:"fingerprint"`
		Port        int    `json:"port"`
		Protocol    string `json:"protocol"`
		Download    bool   `json:"download"`
	} `json:"info"`
	Files map[string]FileMetadata `json:"files"`
}

type UploadMetadataResponseMessage struct {
	SessionID string            `json:"sessionId"`
	Files     map[string]string `json:"files"`
}
