# Relay Node For Qitmeer Network

Qitmeer Relay node is a kind of node which uses circuit transport protocol. 
It can route communication between two peers. So it can also be used to help 
private network qitmeer nodes to "Hole-punching".


Relay node connections are end-to-end encrypted, which means that the peer acting as the relay is unable to read or tamper with any traffic that flows through the connection.


## Installation
```bash
~ cd ./cmd/relaynode
~ go build
~ ./relaynode -h
```
## Getting Started

* Start `relaynode`
```bash
~ ./relaynode
```

* Copy `Relay Address` from relaynode log.

* Configure startup parameters for Qitmeer nodes in the private network.
```bash
~ ./qitmeerd --relaynode=[Relay Address]
```

## Usage

* If you do not want to use the relay node default configuration parameters, you can use `./relaynode -h` to help for custom configuration.

* If your environment is having trouble getting public IP, please try using `./relaynode --externalip=[Your Public IP]`