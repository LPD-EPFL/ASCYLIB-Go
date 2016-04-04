# Build a test binary

*You will need 'gccgo' and 'make' to build a test binary.*

To build a test binary (= *test code* + *data structure* to test), in **src/**, run:

    make build NAME=<data structure name> [ TEST=<test module name> ]

The TEST parameter is optional (don't put brackets!).
The *simple* test module is selected by default.

The list of *test modules* can be found in **src/test/**.
The list of *data structures* can be found in **src/**.

# Run a test binary

Still from **src/**, run:

    make run NAME=<data structure name> [ TEST=<test module name> ]
    
Or from **bin/**, run:

    ./<data structure name>_<test module name>
