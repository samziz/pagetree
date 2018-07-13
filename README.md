# Pagetree


### Description

I designed Pagetree to work like the Unix tree utility. It takes a URL (mandatory) and a list of flags (optional) and returns a map of the site in a tree format. The tree starts at the URL entered and prints all the links it finds on that site.


### Building Pagetree

The easiest way to build this is by using the makefile. Running `make` will install the binary in your `/usr/local/bin` directory so you can run it from the shell with `pagetree ...`. If you don't want to do this, or if you're using a non-Unix system like Windows, you can run `make local` which will build the binary in your current directory, which is probably this directory.


### Implementation

#### Crawling

I use a Crawler struct to store configuration (like command-line flags, or defaults hardcoded into the program) as well as process state (visited links, links banned by the site's robots policy, etc). Then we spin up a lot of goroutines to read pointers to uncrawled nodes from a channel and attach these results as new nodes to the tree.

#### Printing

I used a basic recursive algorithm to mimic the tree utility: it works better on larger screens or with shorter links. 