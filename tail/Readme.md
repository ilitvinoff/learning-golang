Usage: tail -flag "configuration string" -flag1 "configuration string"...
Print the last 10 lines of each FILE to standard output.

The basic functionality is the same as the standard tail utility started with '-F' flag.

Added functionality:
1. You can "tail" multiple files at the same time. This way, you can define a "prefix" for each file, 
   which will be printed before the line from the corresponding file.
2. If file - doesn't exist -> wait for it to appear
3. If the file has been deleted / moved -> wait for a new one to appear. When new one appeared, tailing starts from the begingng of the file.
4. You may select directory and define the regular expression(for file name).
   This way, 'tailer' will keep track of the last file that came up that matches the regular expression.

Warning!!!!
You need to describe parameters of each flag with 1 stringline. So, for example you need to tail 2 files, then both files
you'll describe in one string, using semicolon, as delimiter.

   Example:
   ./tail -p "filepath1;prefix for output from file1;filepath2;prefix for output from file2..."

Warning!!!
Do not use semicolon at the end of argument line.


Flag's description:
  -c (Specify full config for each file)
       Set's full config for each file to tail. All files specify in 1 string, using semicolon as delimiter.
       You need to define such parameters for each file:
    	 path - path to file or directory, depends on your will
    	 regex - regular expression for filename(may be empty string, if u set path - as path to file. empty string means: path;;prefix;n)
    	 prefix - prefix to printout before line from file
    	 n: output the last 'n' lines,may be empty string(Must be integer if present!!!)

    	Example: ./tail -c "foo/bar/file1;regexForFile1;prefix1;someinteger;foo/bar/file2;regexForFile2;prefix2;someinteger..."
    	Warning!!!! Do not use semicolon at the end of argument line.
    	
  -n (Number of lines to standard output.)
    	when 'n' is set, it defines to all files, which's parameter 'n' has default value(default - 10) . 'n' represent amount of string to tail from.
    	Warning!!!! Do not use semicolon at the end of argument line.
    
  -p (when u need to tail from the specified file(group of files))
    	File's path to tail.If you want to specify 1 file path, 1 argument is enough - the file path.
    	You may specify prefix to printout it before textline from file, using semicolon as delimiter. Example: tail -p "foo/bar/file;prefix"
    	If you want to specify more then 1 file, you need to define both parameters(file's path;prefix) for each file in one string.
      
    	Example: ./tail -p "foo/bar/file1;prefix1;foo/bar/file2;prefix2;foo/bar/file3;prefix3..."
    	Warning!!!! Do not use semicolon at the end of argument line.
    	
  -r (when you select a directory where the tailer will keep track of the file matching the regex you defined)
    	Filename regular expression pattern. If you want to specify 1 file to tail, 2 arguments are enough - "pathToDirectory;regex".
    	If you want to specify more then 1 file, you need to define 3 arguments for each file:
      
    	Example: ./tail -r "foo/bar/file1;regexForFile1;prefix1;foo/bar/file2;regexForFile2;prefix2;foo/bar/file3;regexForFile3;prefix3..."
    	Warning!!!! Do not use semicolon at the end of argument line.


Warning!!!
Do not use semicolon at the end of argument line.

