module github.com/wostzone/owserver-pb

go 1.14

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubclient-go v0.0.0-00010101000000-000000000000
	github.com/wostzone/hubserve-go v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200930145003-4acb6c075d10 // indirect
	golang.org/x/sys v0.0.0-20200929083018-4d22bbb62b3c // indirect
)

// Until libs are stable
replace github.com/wostzone/hubclient-go => ../hubclient-go

replace github.com/wostzone/hubserve-go => ../hubserve-go
