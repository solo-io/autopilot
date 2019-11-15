package v1

// warning: generating API protos requires autopilot repo to live in ${GOPATH}

//go:generate protoc --doc_out=../../docs/content/api/v1 --doc_opt=./docs_template.tmpl,api_v1.md --go_out=${GOPATH}/src/ ./autopilot.proto ./autopilot-operator.proto
