package broadcast

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// LocalSend does not require a specific port or multicast address but instead provides a default configuration.
// Everything can be configured in the app settings if the port / address is somehow unavailable.
// The default multicast group is 224.0.0.0/24 because some Android devices reject any other multicast group.
// Multicast (UDP)
//     Port: 53317
//     Address: 224.0.0.167
// HTTP (TCP)
//     Port: 53317

var unique_fingerprint string
var unique_fingerprint_lock sync.RWMutex

func SetUniqueFingerprint(fingerprint string) {
	unique_fingerprint_lock.Lock()
	defer unique_fingerprint_lock.Unlock()
	unique_fingerprint = fingerprint
}

// Announcement
func SendAnnouncement(alias string) (bool, error) {
	if unique_fingerprint == "" {
		return false, fmt.Errorf("unique fingerprint is not set")
	}
	// Send the announcement
	announcement_message := &AnnouncementMessage{
		Alias:       alias,
		Version:     "2.0",
		DeviceModel: "go-lsendd",
		DeviceType:  "tablet",
		Fingerprint: unique_fingerprint,
		Port:        53317,
		Protocol:    "http",
		Download:    false,
		Announce:    false,
	}

	listenAddr, err := net.ResolveUDPAddr("udp4", "224.0.0.167:53317")
	if err != nil {
		return false, fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		return false, fmt.Errorf("failed to listen on UDP address: %v", err)
	}
	defer conn.Close()

	// JSON-ify the announcement message
	jsonData, err := json.Marshal(announcement_message)
	if err != nil {
		return false, fmt.Errorf("failed to marshal announcement message: %v", err)
	}

	// Send the announcement message
	_, err = conn.WriteToUDP(jsonData, listenAddr)
	if err != nil {
		return false, fmt.Errorf("failed to send announcement message: %v", err)
	}

	return true, nil
}
