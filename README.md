# |coin> 
## About
ketcoin is a toy blockchain prototype based on a post-quantum cryptosystem 
developed as a project to better understand the inner workings of blockchain 
systems, peer-to-peer communication and post-quantum cryptography. 
This project was realised within the frame of the course IFT4055 and a written
report can be found [here](https://github.com/tremblaythibaultl/IFT4055/blob/main/rapport.pdf).

This project is authored by Louis Tremblay Thibault and distributed under the 
MIT license.

The cryptographic library developed for this project can be found [here](https://github.com/tremblaythibaultl/AMSS).

## Usage
One is free to download and play around with the code. To do so, you can do 
the following : 

`$ git clone https://github.com/tremblaythibaultl/ketcoin.git`  
`$ cd ketcoin`  
`$ go build`  
`$ ./src -l 13337 -t ketcoin.louistt.com:13337`


This last line is not guaranteed to work all the time (as the seed node will 
not be maintained at all times).

*This code is neither thread-safe nor secure as of right now.*
