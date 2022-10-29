run:
	go run main.go

build:
	go build -o recorder -ldflags "-s -w" main.go

tags::
	gotags -R . > tags
