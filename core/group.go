package core

type Group struct {
	GroupName string   `json:"group_name"`
	Prefix    string   `json:"prefix"`
	Servers   []Server `json:"servers"`
}
