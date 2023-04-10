README.md: *.go
	goreadme --functions --methods --types --variabless --badge-godoc --badge-goreportcard --import-path github.com/anglo-korean/external-metrics > $@
