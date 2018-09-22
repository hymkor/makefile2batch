makefile2batch.exe : main.go
	go fmt
	go build -o $@ -ldflags "-s -w"

test:
	makefile2batch > make.cmd

clean:
	if exist make.cmd del make.cmd
	if exist makefile2batch.exe del makefile2batch.exe

