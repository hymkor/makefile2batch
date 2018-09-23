makefile2batch
==============

Makefile to Batchfile converter.

```
$ makefile2batch [-f Makefile] > make.cmd
```

Supported Macros
----------------

* `$@` ... target filename
* `$*` ... target filename without extension
* `$<` ... first source filename
* `$^` ... all source filenames
* `$(xxxx)` ... the environment variable

Sample
-------

### Source

```Makefile
makefile2batch.exe : main.go
	go fmt
	go build -o $@ -ldflags "-s -w"

test:
	makefile2batch > make.cmd

clean:
	if exist make.cmd del make.cmd
	if exist makefile2batch.exe del makefile2batch.exe
```

### make.cmd created by `makefile2batch > make.cmd`

```make.cmd
@echo off
setlocal
set "MAKEDIR=%CD%"
call :"%1"
endlocal
exit /b
:""
  call :"makefile2batch.exe"
  exit /b

:"makefile2batch.exe"
  call :test makefile2batch.exe main.go
  if not errorlevel 1 exit /b
  echo go fmt
  go fmt
  echo go build -o makefile2batch.exe -ldflags ^"-s -w^"
  go build -o makefile2batch.exe -ldflags "-s -w"
  exit /b

:"test"
  echo makefile2batch ^> make.cmd
  makefile2batch > make.cmd
  exit /b

:"clean"
  echo if exist make.cmd del make.cmd
  if exist make.cmd del make.cmd
  echo if exist makefile2batch.exe del makefile2batch.exe
  if exist makefile2batch.exe del makefile2batch.exe
  exit /b

:test
  if not exist "%~1" exit /b 1
  if "%~2" == "" exit /b 1
  setlocal
  for /F "tokens=2,3" %%I in ('where /R . /T "%~1"') do set TARGET=%%I_%%J
  echo %TARGET% | findstr _[0-9]: > nul && set TARGET=%TARGET:_=_0%

:each_source
  for /F "tokens=2,3" %%I in ('where /R . /T "%~2"') do set SOURCE=%%I_%%J
  echo %SOURCE% | findstr _[0-9]: > nul && set SOURCE=%SOURCE:_=_0%
  if "%SOURCE%" gtr "%TARGET%" exit /b 1
  shift
  if not "%~2" == "" goto :each_source
  endlocal & exit /b 0
```
