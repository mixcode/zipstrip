# zipstrip: a small tool to remove garbage data attached at the end of a ZIP file

## Usage
```
# truncate a zip file in-place without making a backup
$ zipstrip -t -nobackup SOME_ZIP_FILE.zip

# make a new zip without garbage data at the end
$ zipstrip -c SOME_ZIP_FILE.zip NEW_ZIP_FILE.zip

# help
$ zipstrip -help
```

