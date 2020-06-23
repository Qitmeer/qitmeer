# Block and Transaction Broadcasting with ZeroMQ

[ZeroMQ](http://zeromq.org/) is a lightweight wrapper around TCP
connections, inter-process communication, and shared-memory,
providing various message-oriented semantics such as publish/subscribe,
request/reply, and push/pull.

The Qitmeer daemon can be configured to act as a trusted "border
router", implementing the Qitmeer protocol and relay, making
consensus decisions, maintaining the local blockchain database,
broadcasting locally generated transactions into the network, and
providing a queryable RPC interface to interact on a polled basis for
requesting blockchain related data. However, there exists only a
limited service to notify external software of events like the arrival
of new blocks or transactions.

The ZeroMQ facility implements a notification interface through a set
of specific notifiers. Currently there are notifiers that publish
blocks and transactions. This read-only facility requires only the
connection of a corresponding ZeroMQ subscriber port in receiving
software; it is not authenticated nor is there any two-way protocol
involvement. Therefore, subscribers should validate the received data
since it may be out of date, incomplete or even invalid.

ZeroMQ sockets are self-connecting and self-healing; that is,
connections made between two endpoints will be automatically restored
after an outage, and either end may be freely started or stopped in
any order.

Because ZeroMQ is message oriented, subscribers receive transactions
and blocks all-at-once and do not need to implement any sort of
buffering or reassembly.

## Prerequisites
If you want to open it, you must install some dependency libraries:

* Mac:
```
    brew install czmq libsodium
```
* Ubuntu:
```
    apt install czmq libsodium
```
## Enabling

By default, the ZeroMQ is disable.  To enable, use ZMQ=TRUE
during go building qitmmer:

    $ make ZMQ=TRUE

To actually enable operation, one must set the appropriate options on
the command line or in the configuration file.

## Usage

Currently, the default configuration is as follows:
```
    --zmqpubhashtx=*
    --zmqpubhashblock=*
    --zmqpubrawblock=*
    --zmqpubrawtx=*
```
or:
```
    --zmqpubhashtx=default
    --zmqpubhashblock=default
    --zmqpubrawblock=default
    --zmqpubrawtx=default
```
The default detailed address can be found in the log.
Of course, if you need a special address, you can configure it as follows:

```
    --zmqpubhashtx=address
    --zmqpubhashblock=address
    --zmqpubrawblock=address
    --zmqpubrawtx=address
```
