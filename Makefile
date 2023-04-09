APPLICATION := {{ .Name }}
TYPE := consumer

include .build-tools/Makefile

$(dist)/template.yaml: export REPO = github.com/anglo-korean/{{ .Name }}
$(dist)/template.yaml: export RUNBOOKS = [github.com/anglo-korean/{{ .Name }}]
$(dist)/template.yaml: export CONFIG = {"KAFKA_BROKER": "kafka:9092", "VERSION": "$(VERSION)"}
$(dist)/template.yaml: export FAMILY = low
$(dist)/template.yaml: export READS_FROM = {{ .Name }}-incoming
$(dist)/template.yaml: export WRITES_TO = {{ .Name }}-outgoing

vendor/: go.sum
	go mod vendor

docker-build: vendor/
