# asthum #

### Simple HTTP Server ###

~~Subtitle says what it does. Mostly.~~

I have no idea what it does. You have no idea what it does. Nobodies
happy.

It makes usage of go text/templates. At the moment it is in a state of
flux and there is little purpose in me telling you how to use it.

### Name ###

Asthum comes from 'A HTML Web Server That Uses Markdown' -> 'ahwstum' ->
'asthum' : I'm horrible at coming up with names.

The name seems to have quickly become defunct.

### A Basic Idea Of What It Does ###

When a client requests a file, asthum checks the most relevant `.rules` 
file and decideds the file needs to be interpreted by a program, then
if it should use a template.

`.rules` is used to figure out whether to use a template and
whether to return the file as is or to interpret it with a program
first. It has a format like this:

	pattern [templated] [interpreter] [args...]
	pattern hidden

So for example if you wanted to use css and markdown files you would do
something like this:

	.*\.md templated markdown
	.*\.py python
	\.git hidden

When files are interpreted query strings are used to set values in the
environment. Similar to CGI scripts.

The template files executed with a struct like this:

	Name string 
	Link string 
	Content string

Hopefully you now know a little about how to use it.

Check `-h` for arguments.

