# asthum #

### A HTML Web Server That Uses Markdown ###

Title says what it does. Mostly.

Expects files to be markdown, will convert them to html when they are requested with a markdown processor. Defaults to `markdown`, can be changed at runtime with `-m`.

If the requested file ends in '.md' the raw markdown file will be returned without processing.

When processing a file or directory it will search upwards from the file looking for the first occurance of a 'DIR.tmpl' or 'PAGE.tmpl' file, they should be go text/templates. When exectuted they are given a struct with the following fields:

    Link
    Name
    Content    # Parsed content of the file.
    Links      # Map of file names to links in the directory.

Link and Name are avaliable in both DIR and PAGE templates, Links only in DIR and Content only in PAGE.

If a directory is asked for one of three things can happen, the first is that the directory contains a file that has the prefix `index`, if that occures and the file is executable then it will be executed and it's stdout will be given as the response. If it is not executable then it itself will be given as the response. If there is no such file then a list of the files in the directory will be given (with the appropriate template). If there are multiple `index*` files the first is picked, how it orders I know  not.
### Notes ###

Asthum comes from 'A HTML Web Server That Uses Markdown' -> 'ahwstum' -> 'asthum' : I'm horrible at coming up with names.

