# Pagetree


### Description

I designed Pagetree to work like the Unix tree utility. It takes a URL (mandatory) and a list of flags (optional) and returns a map of the site in a tree format. The tree starts at the URL entered and prints all the internal links it finds on that site.

### Building Pagetree

The easiest way to build this is by using the makefile. Running `make` will install the binary in your `/usr/local/bin` directory so you can run it from the shell with `pagetree ...`. If you don't want to do this, or if you're using a non-Unix system like Windows, you can run `make local` which will build the binary in your current directory, which is probably this directory.

### Usage

Assuming you installed Pagetree with `make` (or to somewhere in your execution path), run `pagetree -maxpages 100 https://monzo.com/blog/` substituting any website and any flags. Setting `maxpages` is recommended for very large websites or results can easily get into the >10000 range. The flags are as follows:

```
	-crawlrate: The maximum length of time to wait between requests. The default is 1ms - don't modify unless you want to make a cup of tea while you wait.

	-httptimeout: The maximum length of time to wait for any HTTP response in milliseconds. The default is 2 seconds. 

	-maxroutines: The maximum number of goroutines to spin up. There is no default but you can safely leave this empty.

	-maxpages: The maximum number of pages to crawl. The default is -1 (no limit).

	-respectrobots: If true, pagetree will respect the site's robots.txt policy. The default is false. Be careful - many sites ban or severely restrict crawlers.

	-timeout: The maximum length of time a goroutine will wait for another goroutine to register a new node before it sends a kill signal to the main routine, which wraps up the crawler and renders the tree. The default is 5 seconds. There should be no reason to modify this.
```

### Implementation

#### Crawling

I use a Crawler struct to store configuration (like command-line flags, or defaults hardcoded into the program) as well as process state (visited links, links banned by the site's robots policy, etc). Then we spin up a lot of goroutines to read pointers to uncrawled nodes from a channel and attach these results as new nodes to the tree.

#### Printing

I used a basic recursive algorithm to mimic the tree utility: it works better on larger screens or with shorter links. 


### Structure

The project is split up into different files which sometimes relate to different types and methods (e.g. Crawler, Parser, Map...) but other times are just thematic. Lowercased method names indicate functions that are for internal use by that file. Even though of course they are still accessible within the main package, this is meant as a signal that they are only intended to be used by the high-level interfaces in that file.