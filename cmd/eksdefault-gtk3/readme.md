awsdefault-gtk3 UI
==================

# Usage

```bash
$ awsdefault-gtk3
```

this will start the UI. Each click will update the credentials file and close the application. ![awsdefault-gkt3-example1](../../doc/awsdefault-gtk3-example1.gif?raw=true)

*Note* [i3block](https://github.com/vivien/i3blocks) was used for the status bar in this example. You can find the config in the [doc folder](doc/i3block-example.conf). You will also need to install the cli version of awsdefault, for updating the text in the i3block.

If you need this window permanently open, add the parameter `-permanent` and run:

```bash
$ awsdefault-gtk3 -permanent
```





# Installation

## Option 1 — Download binaries

precompiled binaries for Linux, Windows and MacOS are available at the [release] page.


```bash
curl 
```

## Option 2 — Compile it


### Install the dependencies

- *[Go](https://golang.org/doc/install)* is required
- [gtk3](https://www.gtk.org/) is required
- clone this repository: 

```bash
$ go get github.com/peterbueschel/awsdefault
```

- [go-ini](https://github.com/go-ini/ini); used for the handling of the [AWS credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html),

```bash
$ go get github.com/go-ini/ini
```

- [gotk3](https://github.com/gotk3/gotk3); for the UI,

*Note* if you get error messages like `could not determine kind of name for ...`, check this [issue report](https://github.com/gotk3/gotk3/issues/152)

```bash
$ go get github.com/gotk3/gotk3/gtk
```

- _optional_ [godebug](https://github.com/kylelemons/godebug/pretty); for testing via `go test ./...`

```bash
$ go get github.com/kylelemons/godebug/pretty
```


### Install the Go binary

#### Linux 

```bash
$ cd $GOPATH/src/github.com/peterbueschel/awsdefault/cmd/awsdefault-gtk3/ && go install
```

*if everything went well, the binary can now be found in the directory* _$GOPATH/bin_ 


## Configure your environment (only first time)

Set the environment variable `AWS_PROFILE` to `default` ([aws userguide](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html)).

### Linux

Add the following line to your .xinitrc, .zshrc or .bashrc file:

```bash
export AWS_PROFILE=default
```

### Windows

TODO
