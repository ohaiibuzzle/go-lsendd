package broadcast

type AnnouncementMessage struct {
	Alias       string `json:"alias"`
	Version     string `json:"version"`
	DeviceModel string `json:"device_model"`
	DeviceType  string `json:"device_type"`
	Fingerprint string `json:"fingerprint"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Download    bool   `json:"download"`
	Announce    bool   `json:"announce"`
}
