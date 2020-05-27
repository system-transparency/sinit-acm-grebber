# sinit-acm-grebber

#### simple tool for Downloading and unzipping intel-txt binaries

You can simply run:

```
   $ sinit-acm-grebber
```
for downloading all sinit files from [Intel]( https://software.intel.com/content/www/us/en/develop/articles/intel-trusted-execution-technology.html).


in case the Url has changed you can also use: 
```
   $ sinit-acm-grebber -url https://some.url/at/intel
```


the default output folder ist the current Directory it can be modified can be by using:
```
   $ sinit-acm-grebber -of ./some/directory
```


the zip files can be kept in a sub folder /zip by using:
```
   $ sinit-acm-grebber -noClean
```
