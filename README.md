<div align="center">
	<h1>Got.</h1>
	<h4 align="center">
		Simple and fast concurrent downloader - forked from <a href='https://github.com/melbahja/got'>melbahja/got</a>)
	</h4>
</div>

## Comparison

Comparison in cloud server:

```bash

[root@centos-nyc-12 ~]# time got -o /tmp/test -c 20 https://proof.ovh.net/files/1Gb.dat
URL: https://proof.ovh.net/files/1Gb.dat done!

real    0m8.832s
user    0m0.203s
sys 0m3.176s


[root@centos-nyc-12 ~]# time curl https://proof.ovh.net/files/1Gb.dat --output /tmp/test1
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
								 Dload  Upload   Total   Spent    Left  Speed
100 1024M  100 1024M    0     0  35.6M      0  0:00:28  0:00:28 --:--:-- 34.4M

real    0m28.781s
user    0m0.379s
sys 0m1.970s

```


## Installation

#### Build from source
```bash
mkdir -p $HOME/go/src/github.com/0pcom
cd $HOME/go/src/github.com/0pcom
git clone https://github.com/0pcom/got
cd got
go build cmd/got/got.go
```

#### Run from source
```bash
mkdir -p $HOME/go/src/github.com/0pcom
cd $HOME/go/src/github.com/0pcom
git clone https://github.com/0pcom/got
cd got
go run cmd/got/got.go --help
```

## How It Works?

Got takes advantage of the HTTP range requests support in servers [RFC 7233](https://tools.ietf.org/html/rfc7233), if the server supports partial content Got split the file into chunks, then starts downloading and merging the chunks into the destinaton file concurrently.
