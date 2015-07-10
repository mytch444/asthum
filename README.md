# asthum #

### Simple Http Server That Converts Files ###

~~Subtitle says what it does. Mostly.~~

I have no idea what it does. You have no idea what it does. Nobodies happy.

It makes useage of go text/templates. At the moment it is in a state of flux and there is little purpose in me telling you how to use it.

### Name ###

Asthum comes from 'A HTML Web Server That Uses Markdown' -> 'ahwstum' -> 'asthum' : I'm horrible at coming up with names.

### A Basic Idea Of What It Does ###

Returns files or runs scripts that are requested. Also uses go `text/templates` to make coding a little less repedative.

Files that begin with periods cannot be requested. `asthum` also has three special files names that it uses. These file file used is the first file that it find that matches the name when it looks up the directory tree starting from the path of the file requested. I hope that makes sense. These files are called:

    .page.tmpl
    .interpreters

`.page.tmpl` is a template files that is (sometimes) used to template requested files.

`.interpreters` is used to figure out whether to use a template and whether to return the file as is or to interpret it with a program first. It has a format like this:

    filesuffix [yes|no] [interpreter] [args...]

So for example if you wanted to use css and markdown files you would do something like this:

    md yes markdown
    css no

There is an example file in `test-site/.interpreters`. In fact, you should probably just look in `test-site` for examples of everything.

The template files executed with a struct like this:

    Name string
    Link string
    Content string

Hopefully you now know a little about how to use it.

Check `-h` for arguments.
 