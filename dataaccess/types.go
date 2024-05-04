package dataaccess

import "time"

type elementIdDao struct {
	Id uint64 `sql:"id"`
}

type secretValueDao struct {
	Value string `sql:"value"`
}

type homeAuthCodesDao struct {
	Id       uint64    `sql:"id"`
	ClientId uint64    `sql:"client_id"`
	UserId   uint64    `sql:"user_id"`
	Expire   time.Time `sql:"expire"`
	Code     string    `sql:"code"`
}

type homeAuthDao struct {
	Id    uint64 `sql:"id"`
	Login string `sql:"login"`
	Hash  string `sql:"hash"`
}

type homeClientsDao struct {
	Id                uint64 `sql:"id"`
	ClientId          string `sql:"client_id"`
	ClientSecret      string `sql:"client_secret"`
	ClientRedirectUri string `sql:"client_redirect_uri"`
}

type homeDevicesDao struct {
	Id          uint64 `sql:"id"`
	UserId      uint64 `sql:"user_id"`
	RelayId     uint64 `sql:"relay_id"`
	SwitchId    uint64 `sql:"switch_id"`
	On          bool   `sql:"on"`
	DefaultName string `sql:"default_name"`
	Type        string `sql:"type"`
}

type homeDeviceTasksDao struct {
	Id         uint64    `sql:"id"`
	DeviceId   uint64    `sql:"device_id"`
	Capability string    `sql:"capability"`
	Instance   string    `sql:"instance"`
	Value      string    `sql:"value"`
	CreatedAt  time.Time `sql:"created_at"`
	ExpiresAt  time.Time `sql:"expires_at"`
	UserId     uint64    `sql:"user_id"`
}
