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

readme:
	gawk "/^```make.cmd/{ print $0 ; while( getline < \"make.cmd\" ){ print } ; print \"```\" ; exit } ; 1" readme.md | nkf32 -Lu > readme.new && move readme.new readme.md

clean:
	if exist make.cmd del make.cmd
	if exist makefile2batch.exe del makefile2batch.exe
```

### make.cmd created by `makefile2batch > make.cmd`

```make.cmd
@echo off
rem ***
rem *** Code generated by `makefile2batch`; DO NOT EDIT.
rem *** ( https://github.com/zetamatta/makefile2batch )
rem ***
setlocal
set "PROMPT=$$ "
set "MAKEDIR=%CD%"
call :"%1"
endlocal
exit /b
:""
  call :"makefile2batch.exe"
  exit /b

:"makefile2batch.exe"
  call :test makefile2batch.exe main.go && exit /b
  @echo on
  go fmt
  go build -o makefile2batch.exe -ldflags "-s -w"
  @echo off
  exit /b

:"test"
  @echo on
  makefile2batch > make.cmd
  @echo off
  exit /b

:"readme"
  @echo on
  gawk "/^```make.cmd/{ print $0 ; while( getline < \"make.cmd\" ){ print } ; print \"```\" ; exit } ; 1" readme.md | nkf32 -Lu > readme.new && move readme.new readme.md
  @echo off
  exit /b

:"clean"
  @echo on
  if exist make.cmd del make.cmd
  if exist makefile2batch.exe del makefile2batch.exe
  @echo off
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
