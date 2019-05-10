package app

type IndexType int

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

	termWidth  int
	termHeight int
}

type Operation struct {
	Key     string
	Label   string
	End     bool
	Process func()
}
