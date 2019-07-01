package app

type IndexType int

type LogMode string

const (
	LogModeCover  LogMode = "cover"
	LogModeAppend LogMode = "append"
)

type Config struct {
	ShowDetail bool                   `json:"show_detail"`
	Servers    []Server               `json:"servers"`
	Groups     []Group                `json:"groups"`
	Options    map[string]interface{} `json:"options"`
}

type Group struct {
	GroupName string   `json:"group_name"`
	Prefix    string   `json:"prefix"`
	Servers   []Server `json:"servers"`
	Collapse  bool     `json:"collapse"`
}

type ServerIndex struct {
	indexType   IndexType
	groupIndex  int
	serverIndex int
	server      *Server
}

type Server struct {
	Name     string                 `json:"name"`
	Ip       string                 `json:"ip"`
	Port     int                    `json:"port"`
	User     string                 `json:"user"`
	Password string                 `json:"password"`
	Method   string                 `json:"method"`
	Key      string                 `json:"key"`
	Options  map[string]interface{} `json:"options"`
	Alias    string                 `json:"alias"`
	Log      ServerLog              `json:"log"`

	termWidth  int
	termHeight int
	groupName  string
}

type ServerLog struct {
	Enable   bool    `json:"enable"`
	Filename string  `json:"filename"`
	Mode     LogMode `json:"mode"`
}

type Operation struct {
	Key     string
	Label   string
	End     bool
	Process func(...interface{})
}
