package connector

type Connector interface {
	IsInstalled() (bool, error)
	PreConnect() error
	Connect() error
	PostConnect() error
}

type DefaultConnector struct {
	HostName      string
	Port          int
	UserName      string
	PlainPassword string
	WaitFor       bool
}
