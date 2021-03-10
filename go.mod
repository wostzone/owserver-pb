module github.com/wostzone/owserver

go 1.14

require (
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub v0.0.0-20210305045441-1a35764ba993
)

// Until Hub is stable
replace github.com/wostzone/hub => ../hub
