// internal/fs/types.go
package fs

// Location holds cached geolocation from ipinfo.io.
type Location struct {
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
	Timezone   string `json:"timezone"`
	Loc        string `json:"loc"`
	ResolvedAt string `json:"resolved_at"`
}

// AgentNode represents a discovered agent in the network.
type AgentNode struct {
	Address      string         `json:"address"`
	AgentName    string         `json:"agent_name"`
	Nickname     string         `json:"nickname"`
	State        string         `json:"state"`
	Alive        bool           `json:"alive"`
	IsHuman      bool           `json:"is_human"`
	Capabilities []string       `json:"capabilities"`
	Location     *Location      `json:"location,omitempty"`
	Storage      *StorageStatus `json:"storage,omitempty"`
	WorkingDir   string         `json:"-"` // not serialized to API
}

type StorageStatus struct {
	Schema  string          `json:"schema"`
	Enabled bool            `json:"enabled"`
	Backend string          `json:"backend"`
	Routes  []StorageRoute  `json:"routes"`
	Streams []StorageStream `json:"streams"`
	Health  *StorageHealth  `json:"health,omitempty"`
	NoKV    NoKVStatus      `json:"nokv"`
}

type StorageRoute struct {
	Mount      string `json:"mount"`
	LocalRoot  string `json:"local_root"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
}

type StorageStream struct {
	Stream     string `json:"stream"`
	LocalPath  string `json:"local_path"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
	Mode       string `json:"mode"`
}

type StorageHealth struct {
	Status          string   `json:"status"`
	Backend         string   `json:"backend"`
	Streams         []string `json:"streams"`
	LastError       string   `json:"last_error,omitempty"`
	LastErrorStream string   `json:"last_error_stream,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}

type NoKVStatus struct {
	MetadataAddr string `json:"metadata_addr"`
	Bucket       string `json:"bucket"`
	Endpoint     string `json:"endpoint"`
}

// AvatarEdge is a parent → child spawning relationship.
type AvatarEdge struct {
	Parent    string `json:"parent"`
	Child     string `json:"child"`
	ChildName string `json:"child_name"`
}

// ContactEdge is an agent's declared knowledge of another address.
type ContactEdge struct {
	Owner  string `json:"owner"`
	Target string `json:"target"`
	Name   string `json:"name"`
}

// MailEdge is aggregated communication from sender → recipient.
type MailEdge struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Count     int    `json:"count"`
	Direct    int    `json:"direct"`
	CC        int    `json:"cc"`
	BCC       int    `json:"bcc"`
}

// NetworkStats holds aggregate counts.
type NetworkStats struct {
	Active     int `json:"active"`
	Idle       int `json:"idle"`
	Stuck      int `json:"stuck"`
	Asleep     int `json:"asleep"`
	Suspended  int `json:"suspended"`
	TotalMails int `json:"total_mails"`
}

// Network is the full topology payload returned by the API.
type Network struct {
	Nodes        []AgentNode   `json:"nodes"`
	AvatarEdges  []AvatarEdge  `json:"avatar_edges"`
	ContactEdges []ContactEdge `json:"contact_edges"`
	MailEdges    []MailEdge    `json:"mail_edges"`
	Stats        NetworkStats  `json:"stats"`
	Lang         string        `json:"lang"`
}

// MailMessage is the schema for messages written to mailbox/inbox/{uuid}/message.json.
type MailMessage struct {
	ID          string                 `json:"id"`
	MailboxID   string                 `json:"_mailbox_id"`
	From        string                 `json:"from"`
	To          interface{}            `json:"to"` // string or []string
	CC          []string               `json:"cc"`
	BCC         []string               `json:"bcc"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Type        string                 `json:"type"`
	ReceivedAt  string                 `json:"received_at"`
	SentAt      string                 `json:"sent_at,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
	Identity    map[string]interface{} `json:"identity,omitempty"`

	// Delivered is a transient flag. Not serialized. See TUI mail.go MailCache.
	Delivered bool `json:"-"`
}
