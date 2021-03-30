module github.com/wostzone/owserver

go 1.14

require (
	github.com/iotdomain/iotdomain-go v0.0.0-20200930173842-476b4f672e85
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubapi v0.0.0-00010101000000-000000000000
)

// Until hubapi is stable
replace github.com/wostzone/hubapi => ../hubapi-go
