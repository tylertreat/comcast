#Comcast


Testing distributed systems under hard failures like network partitions and instance termination is critical, but it's also important we test them under [less catastrophic conditions](http://www.bravenewgeek.com/sometimes-kill-9-isnt-enough/) because this is what they most often experience. Comcast is a tool designed to simulate common network problems like latency, bandwidth restrictions, and dropped/reordered/corrupted packets.

It works by wrapping up some system tools in a portable(ish) way. On BSD-derived systems such as OSX, we use tools like `ipfw` and `pfctl` to inject failure. On Linux, we use `iptables` and `tc`. Comcast is merely a thin wrapper around these controls.

## Installation

```
$ go get github.com/tylertreat/comcast
```

## Usage

Currently, Comcast supports just four options: device, latency, bandwidth, and packet loss.

```
$ comcast --device=eth0 --latency=250 --bandwidth=1000 --packet-loss=0.1
```

This will add 250ms of latency, limit bandwidth to 1Mbps, and drop 10% of packets. To turn this off, run the following:

```
$ comcast --mode stop
```

## I don't trust you, this code sucks, I hate Go, etc.

If you don't like running code that executes shell commands for you (despite it being open source, so you can read it and change the code) or want finer-grained control, you can run them directly instead. Read the man pages on these things for more details.

### Linux

On Linux, you can use `iptables` to drop incoming and outgoing packets.

```
$ iptables -A INPUT -m statistic --mode random --probability 0.1 -j DROP
$ iptables -A OUTPUT -m statistic --mode random --probability 0.1 -j DROP
```

Alternatively, you can use `tc` which supports some additional options.

```
$ tc qdisc add dev eth0 root netem delay 50ms 20ms distribution normal
$ tc qdisc change dev eth0 root netem reorder 0.02 duplicate 0.05 corrupt 0.01
```

To reset:

```
$ tc qdisc del dev eth0 root netem
```

### BSD/OSX

To shape traffic in BSD-derived systems, create an `ipfw` pipe and configure it. You can control incoming and outgoing traffic separately as well as which hosts are affected if you want.

```
$ ipfw add 1 pipe 1 ip from me to any
$ ipfw add 2 pipe 1 ip from any to me
$ ipfw pipe 1 config delay 500ms bw 1Mbit/s plr 0.1
```

To reset:

```
$ ipfw delete 1
```

*Note: `ipfw` was removed in OSX Yosemite in favor of `pfctl`.*

## Network Condition Profiles

Here's a list of network conditions to that you can plug into Comcast. Please add any more that you may come across.

Name | Latency | Bandwidth | Packet-loss
:-- | --: | --: | --:
GPRS good | 500 | 50 | 2
EDGE good | 300 | 250 | 1.5
3G(HSDPA) good | 250 | 750 | 1.5
DSL (poor) | 70 | 2000 | 2
DSL (good) | 40 | 8000 | 0.5
WIFI (good) | 40 | 30000 | 0.2
