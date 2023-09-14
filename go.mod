module autossh

go 1.12

require (
	github.com/kr/fs v0.1.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/sftp v1.10.0
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.10.0
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.13.0
	golang.org/x/net => github.com/golang/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190813064441-fde4db37ae7a
)
