makefile2batch
==============

Makefile to Batchfile converter.

```
$ makefile2batch [-f Makefile] {MACRO=VALUE} > make.cmd
```
OR
```
$ makefile2batch [-f Makefile] [-o make.cmd] {MACRO=VALUE}
```

Supported Macros
----------------

* `$@` ... target filename
* `$*` ... target filename without extension
* `$<` ... first source filename
* `$^` ... all source filenames
* `$(MAKE)` ... `CMD /C %~f0`
* `$(MAKEDIR)` ... `%~dp0`
* `$$` ... replace `$`

Sample
-------

### Source

```Makefile
TARGET=makefile2batch.exe

$(TARGET): main.go
	go fmt
	go build -o $@ -ldflags "-s -w"

test:
	makefile2batch > make.cmd

readme:
	gawk "/^```make.cmd/{ print $0 ; while( getline < \"make.cmd\" ){ print } ; print \"```\" ; exit } ; 1" readme.md | nkf32 -Lu > readme.new && move readme.new readme.md

clean:
	if exist make.cmd del make.cmd
	if exist makefile2batch.exe del makefile2batch.exe

upgrade:
	for /F "skip=1" %%I in ('where $(TARGET)') do copy /-Y /v "$(TARGET)" "%%I"
```

### make.cmd created by `makefile2batch > make.cmd`

```make.cmd
@rem ***
@rem *** Code generated by `makefile2batch`; DO NOT EDIT.
@rem *** ( https://github.com/zetamatta/makefile2batch )
@rem ***
@setlocal
@set "PROMPT=$$ "
@call :"%~1"
@endlocal
@exit /b %ERRORLEVEL%
:""
  @call :"makefile2batch.exe"
  @exit /b %ERRORLEVEL%
:errpt
  @echo ERROR %ERRORLEVEL%
  @exit /b %ERRORLEVEL%

:"clean"
  @if exist "clean" @echo '%~f0': 'clean' is up to date. & @exit /b
  if exist make.cmd del make.cmd || goto errpt
  if exist makefile2batch.exe del makefile2batch.exe || goto errpt
  @exit /b

:"makefile2batch.exe"
  @call :test makefile2batch.exe main.go && @echo '%~f0': 'makefile2batch.exe' is up to date. & @exit /b
  go fmt || goto errpt
  go build -o makefile2batch.exe -ldflags "-s -w" || goto errpt
  @exit /b

:"readme"
  @if exist "readme" @echo '%~f0': 'readme' is up to date. & @exit /b
  gawk "/^```make.cmd/{ print $0 ; while( getline < \"make.cmd\" ){ print } ; print \"```\" ; exit } ; 1" readme.md | nkf32 -Lu > readme.new && move readme.new readme.md || goto errpt
  @exit /b

:"test"
  @if exist "test" @echo '%~f0': 'test' is up to date. & @exit /b
  makefile2batch > make.cmd || goto errpt
  @exit /b

:"upgrade"
  @if exist "upgrade" @echo '%~f0': 'upgrade' is up to date. & @exit /b
  for /F "skip=1" %%I in ('where makefile2batch.exe') do copy /-Y /v "makefile2batch.exe" "%%I" || goto errpt
  @exit /b

:test
  @if not exist "%~1" @exit /b 1
  @if "%~2" == "" @exit /b 0
  @setlocal
  @for /F "tokens=2,3" %%I in ('where /R . /T "%~1"') do @set TARGET=%%I_%%J
  @echo %TARGET% | findstr _[0-9]: > nul && set TARGET=%TARGET:_=_0%

:each_source
  @for /F "tokens=2,3" %%I in ('where /R . /T "%~2"') do @set SOURCE=%%I_%%J
  @echo %SOURCE% | findstr _[0-9]: > nul && @set SOURCE=%SOURCE:_=_0%
  @if "%SOURCE%" gtr "%TARGET%" @exit /b 1
  @shift
  @if not "%~2" == "" goto :each_source
  @endlocal & @exit /b 0
```
