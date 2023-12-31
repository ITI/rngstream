# RngStream

[![Report Card](https://goreportcard.com/badge/github.com/iti/rngstream)](https://goreportcard.com/report/github.com/iti/rngstream) 
[![Go Reference](https://pkg.go.dev/badge/github.com/iti/rngstream.svg)](https://pkg.go.dev/github.com/iti/rngstream) 
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Coverage](https://img.shields.io/badge/Coverage-86.9%25-brightgreen)

Package rngstream is an object-oriented random-number package
with many long streams and substreams, based on the
MRG32k3a RNG from reference [1] below and proposed in [2].

It has implementations in
[C](https://www-labs.iro.umontreal.ca/~lecuyer/myftp/streams00/c/),
[C++](https://www-labs.iro.umontreal.ca/~lecuyer/myftp/streams00/c++/),
Go,
[Java](https://www.iro.umontreal.ca/~lecuyer/myftp/streams00/java/),
[R](https://www.rdocumentation.org/packages/parallel/versions/3.6.2/topics/RNGstreams),
OpenCL, and some other
languages.  The main description and documentation is in the
[c++ package](http://www.iro.umontreal.ca/~lecuyer/myftp/streams00/c++/),
in the paper
[streams4.pdf](http://www.iro.umontreal.ca/~lecuyer/myftp/streams00/c++/streams4.pdf).  
The implementations for
[c](http://www.iro.umontreal.ca/~lecuyer/myftp/streams00/c/) and
[java](http://www.iro.umontreal.ca/~lecuyer/myftp/streams00/java/)
give a short description of the interfaces in C and Java, respectively.

The package is copyrighted by Pierre L'Ecuyer and the University of Montreal.
It can be used freely for any purpose.  

e-mail:  lecuyer@iro.umontreal.ca
http://www.iro.umontreal.ca/~lecuyer/

If you use it for your research, please cite the following relevant publications in which MRG32k3a 
and the package with multiple streams were proposed:

[1](https://www-labs.iro.umontreal.ca/~lecuyer/myftp/papers/opres-combmrg2-1999.pdf)
P. L'Ecuyer,
``Good Parameter Sets for Combined Multiple Recursive Random
Number Generators'', Operations Research, 47, 1 (1999), 159--164.

[2](https://www-labs.iro.umontreal.ca/~lecuyer/myftp/papers/streams00.pdf) P. L'Ecuyer, R. Simard, E. J. Chen, and W. D. Kelton, 
``An Objected-Oriented Random-Number Package with Many Long Streams
and Substreams'',
Operations Research, 50, 6 (2002), 1073--1075

Thank you.

(The above text modified from http://www.iro.umontreal.ca/~lecuyer/myftp/streams00/readme.txt).

This Go translation is copyright 2023 The Board of Trustees of the
University of Illinois. See [LICENSE.md](LICENSE.md) for details.
