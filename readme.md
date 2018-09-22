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
