package ecosmodel

type ECoSConnectionInput struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ECoSConnectionResult struct {
	Connected           bool                    `json:"connected"`
	Host                string                  `json:"host"`
	Port                int                     `json:"port"`
	Status              string                  `json:"status,omitempty"`
	ProtocolVersion     string                  `json:"protocolVersion,omitempty"`
	ApplicationVersion  string                  `json:"applicationVersion,omitempty"`
	HardwareVersion     string                  `json:"hardwareVersion,omitempty"`
	CompatibilityStatus ECoSCompatibilityStatus `json:"compatibilityStatus"`
	MissingFields       []string                `json:"missingFields"`
	Message             string                  `json:"message"`
	RawLines            []string                `json:"rawLines,omitempty"`
	Fields              map[string]string       `json:"fields,omitempty"`
}

type ECoSCompatibilityStatus string

const (
	ECoSCompatibilityUnreachable ECoSCompatibilityStatus = "unreachable"
	ECoSCompatibilityRejected    ECoSCompatibilityStatus = "rejected"
	ECoSCompatibilityPartial     ECoSCompatibilityStatus = "partial"
	ECoSCompatibilityUnverified  ECoSCompatibilityStatus = "unverified"
	ECoSCompatibilityVerified    ECoSCompatibilityStatus = "verified"
)

type ECoSLocomotive struct {
	ObjectID    int                 `json:"objectId"`
	Name        string              `json:"name,omitempty"`
	Address     int                 `json:"address"`
	Protocol    string              `json:"protocol,omitempty"`
	Profile     string              `json:"profile,omitempty"`
	Functions   []ECoSFunction      `json:"functions,omitempty"`
	Attributes  map[string][]string `json:"attributes,omitempty"`
	DetailError string              `json:"detailError,omitempty"`
}

type ECoSFunction struct {
	Index       int `json:"index"`
	Description int `json:"description,omitempty"`
}

type Argument struct {
	Key    string
	Value  string
	Params []string
}

type ECoSRawProbe struct {
	Host        string              `json:"host"`
	Port        int                 `json:"port"`
	ProbeFields []string            `json:"probeFields"`
	Locomotives []ECoSRawLocomotive `json:"locomotives"`
	RawLines    []string            `json:"rawLines,omitempty"`
	Message     string              `json:"message"`
}

type ECoSLocomotiveSummary struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Count   int    `json:"count"`
	Message string `json:"message"`
}

type ECoSRawLocomotive struct {
	ObjectID          int                   `json:"objectId"`
	Name              string                `json:"name,omitempty"`
	Address           int                   `json:"address,omitempty"`
	Protocol          string                `json:"protocol,omitempty"`
	Profile           string                `json:"profile,omitempty"`
	Functions         []ECoSFunction        `json:"functions,omitempty"`
	CVs               []ECoSCVValue         `json:"cvs,omitempty"`
	Attributes        map[string][]string   `json:"attributes,omitempty"`
	SupportedFields   []string              `json:"supportedFields,omitempty"`
	MissingFields     []string              `json:"missingFields,omitempty"`
	InterestingFields []string              `json:"interestingFields,omitempty"`
	Probes            []ECoSRawCommandProbe `json:"probes,omitempty"`
	DetailError       string                `json:"detailError,omitempty"`
	CVRead            *ECoSCVReadState      `json:"cvRead,omitempty"`
	CVSources         *ECoSCVSourceInfo     `json:"cvSources,omitempty"`
	FunctionSnapshot  *ECoSFunctionSnapshot `json:"functionSnapshot,omitempty"`
}

type ECoSRawCommandProbe struct {
	Command    string              `json:"command"`
	Fields     []string            `json:"fields"`
	OK         bool                `json:"ok"`
	Status     string              `json:"status,omitempty"`
	Error      string              `json:"error,omitempty"`
	RawLines   []string            `json:"rawLines,omitempty"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

type ECoSCVValue struct {
	Number    int  `json:"number"`
	Value     int  `json:"value"`
	IndexCV31 *int `json:"indexCV31,omitempty"`
	IndexCV32 *int `json:"indexCV32,omitempty"`
}

type ECoSLiveMonitorState string

const (
	ECoSLiveStopped     ECoSLiveMonitorState = "stopped"
	ECoSLiveRunning     ECoSLiveMonitorState = "running"
	ECoSLiveInterrupted ECoSLiveMonitorState = "interrupted"
)

type ECoSLivePulseSample struct {
	At               string `json:"at"`
	RepliesPerSecond int    `json:"repliesPerSecond"`
}

type ECoSLiveEvent struct {
	At       string `json:"at"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Protocol string `json:"protocol"`
}

type ECoSLiveDiagnosis struct {
	ConnectionState             ECoSLiveMonitorState `json:"connectionState"`
	LastSuccessfulCommunication string               `json:"lastSuccessfulCommunication,omitempty"`
	LastError                   string               `json:"lastError,omitempty"`
	Passive                     bool                 `json:"passive"`
}

type ECoSLiveStatus struct {
	Provider             string                `json:"provider"`
	Connected            bool                  `json:"connected"`
	State                ECoSLiveMonitorState  `json:"state"`
	Host                 string                `json:"host,omitempty"`
	Port                 int                   `json:"port,omitempty"`
	StartedAt            string                `json:"startedAt,omitempty"`
	LastSeenAt           string                `json:"lastSeenAt,omitempty"`
	LastMessage          string                `json:"lastMessage,omitempty"`
	BlocksReceived       int                   `json:"blocksReceived"`
	RepliesReceived      int                   `json:"repliesReceived"`
	EventsReceived       int                   `json:"eventsReceived"`
	SubscriptionCommands []string              `json:"subscriptionCommands,omitempty"`
	PulseSamples         []ECoSLivePulseSample `json:"pulseSamples"`
	RecentEvents         []ECoSLiveEvent       `json:"recentEvents"`
	Diagnosis            ECoSLiveDiagnosis     `json:"diagnosis"`
	Error                string                `json:"error,omitempty"`
	Message              string                `json:"message"`
}

type ECoSLocomotiveSyncInput struct {
	Host     string                    `json:"host"`
	Port     int                       `json:"port"`
	ObjectID int                       `json:"objectId"`
	Desired  ECoSLocomotiveSyncDesired `json:"desired"`
	DryRun   bool                      `json:"dryRun"`
	Confirm  bool                      `json:"confirm"`
}

type ECoSLocomotiveSyncDesired struct {
	Name     string `json:"name,omitempty"`
	Address  int    `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type ECoSLocomotiveSyncSnapshot struct {
	Name     string `json:"name,omitempty"`
	Address  int    `json:"address,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type ECoSLocomotiveSyncChange struct {
	Field   string `json:"field"`
	Current string `json:"current"`
	Desired string `json:"desired"`
}

type ECoSLocomotiveSyncResult struct {
	Host     string                     `json:"host"`
	Port     int                        `json:"port"`
	ObjectID int                        `json:"objectId"`
	DryRun   bool                       `json:"dryRun"`
	Applied  bool                       `json:"applied"`
	Current  ECoSLocomotiveSyncSnapshot `json:"current"`
	Desired  ECoSLocomotiveSyncDesired  `json:"desired"`
	Changes  []ECoSLocomotiveSyncChange `json:"changes"`
	Commands []string                   `json:"commands,omitempty"`
	RawLines []string                   `json:"rawLines,omitempty"`
	Message  string                     `json:"message"`
}

type ECoSCVReadState struct {
	Source      string `json:"source,omitempty"`
	NextCV      int    `json:"nextCV"`
	Unavailable int    `json:"unavailable"`
	ReadAt      string `json:"readAt"`
	SourceHost  string `json:"sourceHost"`
}

type ECoSCVSourceConflict struct {
	Number        int    `json:"number"`
	IndexCV31     *int   `json:"indexCV31,omitempty"`
	IndexCV32     *int   `json:"indexCV32,omitempty"`
	DatabaseValue int    `json:"databaseValue"`
	BackupValue   int    `json:"backupValue"`
	BackupSource  string `json:"backupSource,omitempty"`
	RawCacheValue *int   `json:"rawCacheValue,omitempty"`
}

type ECoSCVSourceInfo struct {
	Sources           []string               `json:"sources"`
	BackupStatus      string                 `json:"backupStatus"`
	BackupAdded       int                    `json:"backupAdded"`
	Conflicts         []ECoSCVSourceConflict `json:"conflicts"`
	Warning           string                 `json:"warning,omitempty"`
	StoredCacheStatus string                 `json:"storedCacheStatus,omitempty"`
	UnverifiedValues  []ECoSCVValue          `json:"unverifiedValues,omitempty"`
}

type ECoSFunctionValue struct {
	Index       int  `json:"index"`
	Description int  `json:"description"`
	Momentary   bool `json:"momentary"`
}

type ECoSFunctionSnapshot struct {
	ObjectID  int                 `json:"objectId"`
	Host      string              `json:"host"`
	ReadAt    string              `json:"readAt"`
	Functions []ECoSFunctionValue `json:"functions"`
}
