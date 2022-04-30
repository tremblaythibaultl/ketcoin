# |coin> 
## About
ketcoin is a toy blockchain prototype implementing a post-quantum cryptosystem 
developed as a project to better understand the inner workings of blockchain 
systems, peer-to-peer communication and post-quantum cryptography.
The 

This project authored by Louis Tremblay Thibault and distributed under the 
MIT license.

The cryptographic library developed for this project can be found [here](https://github.com/tremblaythibaultl/AMSS).

For testing purposes, a node will be active from time to time at 
ketcoin.louistt.com:13337.

## Usage
One is free to download and play around with the code. To do so, you can do 
the following : 

`$ git clone https://github.com/tremblaythibaultl/ketcoin.git`  
`$ cd ketcoin`  
`$ go build`  
`$ ./src -l 13337 -t ketcoin.louistt.com:13337`


This last line is not guaranteed to work all the time (as the seed node will 
not be maintained at all times) but you are free to start your own network 
if you wish to.

*This code is not tread-safe and secure to use as of right now.*
