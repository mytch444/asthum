# asthum #

### A HTML Web Server That Uses Markdown ###

Subtitle says what it does. Mostly.

Converts files from markdown with an interpreter (defaults to `markdown`, can be changed with `-m`).

When processing a file or directory it will search upwards from the file looking for the first occurance of a `.PAGE.tmpl` file, which is expected to be a go text/template template. It is given the a struct with the following fields.

    Link
    Name
    Content

They are all strings. Content is html derived from the parsed markdown.

When directories are requested the directory is first searched for a file that begins with index. If it is found then that is returned the same as if that had been requested. If no index file was found then a list of subfiles (files begining with a period are not shown) is created. Then a template named `.DIR.tmpl` is searched for and given a struct with the following fields:

    Link
    Name
    Links

Link and Name are strings, Links is a map of string to string, it's key is the file name and this value is a link to that file.

A special case with files occurs if they are executable (have a x in the mode string, it is not a very good way of checking but I could not think of anything better). If this is the case then the file is executed. If there was any query string in the url requested then the `key=value` pairs are given as arguemnts. The output is returned as the request.


If the requested file name (the last section of the url path) contains a period then the raw file is returned. So if you have '/blog/hi.md' and they request that they get the raw markdown, but if they request '/blog/hi' they get the processed version. This is useful for other file types such as 'html' and tar archives as they will be returned without any processing.

### Notes ###

Asthum comes from 'A HTML Web Server That Uses Markdown' -> 'ahwstum' -> 'asthum' : I'm horrible at coming up with names.

