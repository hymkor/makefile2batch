TARGET=makefile2batch.exe
SHELL=CMD.exe

$(TARGET): main.go
	go fmt
	go build -o $@ -ldflags "-s -w"

test:
	makefile2batch > make.cmd

clean:
	if exist make.cmd del make.cmd
	if exist makefile2batch.exe del makefile2batch.exe

upgrade:
	for /F "skip=1" %%I in ('where $(TARGET)') do copy /-Y /v "$(TARGET)" "%%I"
