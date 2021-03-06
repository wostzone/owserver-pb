module github.com/wostzone/owserver

go 1.14

require (
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/gateway v0.0.0-20210305045441-1a35764ba993
)

// Until gateway is stable
replace github.com/wostzone/gateway => ../gateway
