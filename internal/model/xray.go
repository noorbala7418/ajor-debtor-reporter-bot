package model

type Inbounds struct {
	Inbounds []Inbound `json:"obj"`
}

type Inbound struct {
	Enable   bool     `json:"enable"`
	Remark   string   `json:"remark"`
	Settings string   `json:"settings"`
	Clients  []Client `json:"clientStats"`
}

type Client struct {
	Enable          bool   `json:"enable"`
	AdminEnabled    bool   `json:"-"`
	TotalTraffic    int    `json:"total"`
	UploadTraffic   int    `json:"up"`
	DownloadTraffic int    `json:"down"`
	ID              string `json:"-"`
	Name            string `json:"email"`
}

type Settings struct {
	Clients []SettingClinet `json:"clients"`
}

type SettingClinet struct {
	Enable bool   `json:"enable"`
	Name   string `json:"email"`
	ID     string `json:"id"`
}
