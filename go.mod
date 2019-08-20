module autossh

go 1.12

require (
	github.com/kr/fs v0.1.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/sftp v1.10.0
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/sys v0.0.0-20190509141414-a5b02f93d862 // indirect
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net => github.com/golang/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190813064441-fde4db37ae7a
)
