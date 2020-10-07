check:
	golangci-lint run

check-clean-cache:
	golangci-lint cache clean

protoc-setup:
	wget -P meshes https://raw.githubusercontent.com/layer5io/meshery/master/meshes/meshops.proto

proto:
	protoc -I meshes/ meshes/meshops.proto --go_out=plugins=grpc:./meshes/