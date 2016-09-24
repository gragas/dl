# Overview

There are two ways to approach this problem that I thought of:

1. Split up the download across n routines, giving each routine a range
   of size `$contentLength / n`. Have each routine write into part of a buffer,
   and then write the buffer to disk.
2. Same as 1, but have each routine open up the destination file and write
   into the correct position.

I chose number 1. It's merits are:

* the number of open file descriptors is NOT proportional to the number of goroutines
* it's easier to write the checksum code if we have all the bytes in contiguous memory

The merits of approach 2 are:

* the entire file doesn't have to be in memory at any given time

Since the file to be downloaded is so small, approach 1 seemed like a better option.

# Checking File Integrity

Since the file is stored at `https://storage.googleapis.com/` and it's not
a composite object, I knew that an MD5 sum would be provided in the header
 as either `ETag` or `x-goog-hash`. My program attempts to use `Etag`, and
 if that fails it uses `x-goog-hash`.

# Program Structure

The structure of my program isn't how I would normally structure a large Go project.
It's usually best to separate out the core functionalities into packages.
It's also usually a good idea to avoid making assumptions about how each package
will be used. For example, if this project were a larger one --- a general
download library, say --- then it would be a good idea for the `download` package
to make no assumptions about who is using it: the `download` package should not
print to stdout such that CLI programs built with it look nice. Nor should it
call any other extraneous functions or update a GUI.

But this task was very narrow in scope, so I broke many general rules that keep
larger projects maintainable and scalable.

# Benchmarks

| Number of Routines | Approx. Time Elapsed (seconds) |
| ----------         |                     ---------- |
| wget               |                           8.75 |
| 1                  |                            9.0 |
| 2                  |                            9.4 |
| 4                  |                            9.3 |
| 8                  |                            9.3 |

As I suspected, downloading in parallel does not give any speed
boosts on my setup since the limiting factor seems to be my internet package,
not the number of cores I'm using to download.

# Installation and Use

To install, `go get github.com/gragas/dl`

To run, add $GOBIN to your PATH and run `dl`

You can change the number of goroutines in `main.go`. I did not have enough
time to create a CLI.
