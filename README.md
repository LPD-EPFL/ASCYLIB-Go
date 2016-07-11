ASCYLIB-Go
==========

ASCYLIB-Go is an endeavor to re-implement ASCYLIB + OPTIK (https://github.com/LPD-EPFL/ASCYLIB) in Go.

Only a few of the concurrent-search data-structures implemented in ASCYLIB have been translated.
The philosophy of Go, namely "share memory by communicating; don't communicate by sharing memory", may frown upon such implentation.

ASCYLIB is a set of more than 50 concurrent-search data-structure implementations.
OPTIK is a new design pattern for easily implementing fast and scalable concurrent data structures.

* Website             : http://lpd.epfl.ch/site/ascylib - http://lpd.epfl.ch/site/optik
* Authors             : Vasileios Trigonakis <vasileios.trigonakis@epfl.ch>,
                        Tudor David <tudor.david@epfl.ch>
* Related Publications:
  * *Optimistic Concurrency with OPTIK*,
    Rachid Guerraoui, Vasileios Trigonakis (alphabetical order),
  PPoPP 2016 *(to appear)*
  * *Asynchronized Concurrency: The Secret to Scaling Concurrent Search Data Structures*,
  Tudor David, Rachid Guerraoui, Vasileios Trigonakis (alphabetical order),
  ASPLOS 2015

Algorithms
----------

The following table contains the algorithms translated in ASCYLIB-Go:

| # |    Name                                                                               | Type       | Year | Reference                 |
|:-:|-----------|:-----:|:-----:|:-----:|
|| **Linked lists** ||||
|1|  [Pugh's linked list](./src/linkedlist_pugh.go)                                         | lock-based | 1990 | [[P+90]](#P+90)           |
|2|  [Lazy linked list](./src/linkedlist_lazy.go)                                           | lock-based | 2006 | [[HHL+06]](#HHL+06)       |
|3|  [Harris linked list with ASCY](./src/linkedlist_harris_opt.go)                         | lock-free  | 2015 | [[DGT+15]](#DGT+15)       |
|4|  [OPTIK fine-grained linked list](./src/linkedlist_optik.go)                            | lock-based | 2016 | [[GT+16]](#GT+16)         |
|| **Hash Tables** ||||
|5|  [Java's ConcurrentHashMap](./src/hashtable-java.go)                                    | lock-based | 2003 | [[L+03]](#L+03)           |
|6|  [Hash table using Java's CopyOnWrite array map](./src/hashtable-copy.go)               | lock-based | 2004 | [[ORACLE+04]](#ORACLE+04) |
|7|  [Hash table using global-lock OPTIK list](./src/hashtable-optik1.go)                   | lock-based | 2016 | [[GT+16]](#GT+16)         |
|| **Skip Lists** ||||
|8|  [Sequential skip list](./src/skiplist-seq.go)                                          | sequential |      |                           |
|9|  [Pugh skip list](./src/skiplist-pugh.go)                                               | lock-based | 1990 | [[P+90]](#P+90)           |
|10| [Fraser skip list](./src/skiplist-fraser.go)                                           | lock-free  | 2003 | [[F+03]](#F+03)           |
|11| [Herlihy et al. skip list](./src/skiplist-herlihy_lb.go)                               | lock-based | 2007 | [[HLL+07]](#HLL+07)       |
|12| [OPTIK skip list using trylocks (*default OPTIK skip list*)](./src/skiplist-optik1.go) | lock-based | 2016 | [[GT+16]](#GT+16)         |
|| **Queues** ||||
|13| [Michael and Scott (MS) lock-based queue](./src/queue-ms_lb.go)                        | lock-based | 1996 | [[MS+96]](#MS+96)         |
|14| [Michael and Scott (MS) lock-free queue](./src/queue-ms_lf.go)                         | lock-free  | 1996 | [[MS+96]](#MS+96)         |
|15| [MS queue with OPTIK trylock-version](./src/queue-optik1.go)                           | lock-based | 2016 | [[GT+16]](#GT+16)         |
|16| [MS queue with OPTIK trylock-version](./src/queue-optik2.go)                           | lock-based | 2016 | [[GT+16]](#GT+16)         |
|| **Priority Queues** ||||
|17| [Lotan and Shavit priority queue](./src/priorityqueue-lotanshavit_lf.go)               | lock-free  | 2000 | [[LS+00]](#LS+00)         |
|| **Stacks** ||||
|18| [Global-lock stack](./src/stack-lock.go)                                               | lock-based |      |                           |
|19| [Treiber stack](./src/stack-treiber.go)                                                | lock-free  | 1986 | [[T+86]](#T+86)           |

References
----------

* <a name="DGT+15">**[DGT+15]**</a>
T. David, R. Guerraoui, and V. Trigonakis.
*Asynchronized Concurrency: The Secret to Scaling Concurrent Search Data Structures*.
ASPLOS '15.
* <a name="F+03">**[F+03]**</a>
K. Fraser.
*Practical Lock-Freedom*.
PhD thesis, University of Cambridge, 2004.
* <a name="GT+16">**[GT+16]**</a>
R. Guerraoui, and V. Trigonakis.
*Optimistic Concurrency with OPTIK*.
PPoPP '16.
* <a name="HHL+06">**[HHL+06]**</a>
S. Heller, M. Herlihy, V. Luchangco, M. Moir, W. N. Scherer, and N. Shavit.
*A Lazy Concurrent List-Based Set Algorithm*.
OPODIS '05.
* <a name="HLL+07">**[HLL+07]**</a>
M. Herlihy, Y. Lev, V. Luchangco, and N. Shavit.
*A Simple Optimistic Skiplist Algorithm*.
SIROCCO '07.
* <a name="L+03">**[L+03]**</a>
D. Lea.
*Overview of Package util.concurrent Release 1.3.4*.
http://gee.cs.oswego.edu/dl/classes/EDU/oswego/cs/dl/util/concurrent/intro.html,
2003.
* <a name="LS+00">**[LS+00]**</a>
I. Lotan and N. Shavit.
*Skiplist-based concurrent priority queues*.
IPDPS '00.
* <a name="MS+96">**[MS+96]**</a>
M. M. Michael and M. L. Scott.
*Simple, Fast, and Practical Non-blocking and Blocking Concurrent Queue Algorithms*.
PODC '96.
* <a name="ORACLE+04">**[ORACLE+04]**</a>
Oracle.
*Java CopyOnWriteArrayList*.
http://docs.oracle.com/javase/7/docs/api/java/util/concurrent/CopyOnWriteArrayList.html.
* <a name="P+90">**[P+90]**</a>
W. Pugh.
*Concurrent Maintenance of Skip Lists*.
Technical report, 1990.
* <a name="T+86">**[T+86]**</a>
R. Treiber.
*Systems Programming: Coping with Parallelism*.
Technical report, 1986.

Tests modules
-------------

There is a few test modules available.
All of them execute a concurrent algorithm following variable parameters.

To get the accepted parameters, issue from **src/**:

    make run NAME=<concurrent algorithm name> TEST=simple ARGS=-h

The default test module, 'simple', is the same as in ASCYLIB (https://github.com/LPD-EPFL/ASCYLIB/blob/master/src/tests/test_simple.c).

The 'ldi' test module performs simple latency measurements, for each operation (find, insert, remove).

The three other ones are to get metrics about the Go runtime while performing the same work as the 'simple' test module.
You will need `go tool {trace, pprof}` version 1.6 or higher to build and use those metrics.

Compilation
-----------

### Build a test binary

You will need `gccgo` and `make` to build a test binary.

To build a test binary (= *test code* + *concurrent algorithm* to test), in **src/**, run:

    make build NAME=<concurrent algorithm name> [ TEST=<test module name> ]

The TEST parameter is optional (don't put brackets!).
The 'simple' test module is selected by default.

### Run a test binary

Still from **src/**, run:

    make run NAME=<concurrent algorithm name> [ TEST=<test module name> ARGS=<arguments to pass>]

Or from **bin/**, run:

    ./<concurrent algorithm name>_<test module name>

Thanks
------

Some of the initial implementations used in ASCYLIB were taken from Synchrobench (https://github.com/gramoli/synchrobench -  V. Gramoli. More than You Ever Wanted to Know about Synchronization. PPoPP 2015.).
