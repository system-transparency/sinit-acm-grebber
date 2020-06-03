# sinit-acm-grebber

Simple tool for downloading and unzipping authenticated code modules.

##### Installation

run:
```
   $ go get github.com/system-transparency/sinit-acm-grebber
```
Make sure `${GOPATH}/bin` is in your `$PATH`

##### Usage
You can simply run:

```
   $ sinit-acm-grebber
```
For downloading all SINIT ACMs from [Intel]( https://software.intel.com/content/www/us/en/develop/articles/intel-trusted-execution-technology.html).


In case the URL has changed you can also use: 
```
   $ sinit-acm-grebber -url https://some.url/at/intel
```


The default output folder ist the current directory. It can be modified by using:
```
   $ sinit-acm-grebber -of ./some/directory
```


The zip files can be kept in a sub folder /zip by using:
```
   $ sinit-acm-grebber -noClean
```
