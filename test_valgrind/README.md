# Valgrind Testing for Timing Leaks

This directory contains a test program for the ml-dsa library that, when compiled,
can be tested with Valgrind [to check that functions are constant-time](https://www.imperialviolet.org/2010/04/01/ctgrind.html).

```terminal
go build -o test_program
valgrind --tool=memcheck --track-origins=yes ./test_program
```
