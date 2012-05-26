# vitamix: time virtualization for Go

_vitamix_ is a source-to-source compiler and a runtime for "virtualizing time"
in Go programs. It is intended to aid testing and experimentation of
extremely time-sensitive control software, which is itself written in Go.

## Problem statement

Control software (like robotic control or congestion control) is software that in one
form or another listens to external events in real time and reacts with carefully-timed
responses.

Control software is notoriously hard to get right, because even the computational micro-delays
in the software itself (not to mention the hardware) can affect the overall outcomes.

While developing control software, ideally, one wants to be able to perform
sandbox simulations and experiments, while meeting the following demands:

* _The computational time (i.e. the time it takes the software to compute a reaction to an event) is zero._ 
Being able to do this is a way of decoupling the outcomes of the response logic
(whose timing is of central importance) from the delays incurred by
implementation and hardware peculiarities.

* _Execution outcomes are deterministically exactly reproducible, both in value and in timing._ 
When writing tests for control software, the testing framework often needs to
check whether responses occur at desired times. This is hard to achieve while
using real time since even the outside temperature can affect the execution speed
of your code.

* _Executions that are long-running in real time complete instantaneously in the sandbox._
More often than not, we like to simulate a control algorithm for a long time
(e.g. minutes) to verify that its behavior is stable over time.  A long-running
control program often sleeps most of the time, while waking up at sparse
moments to respond to some external event. In a sandbox, we would really like
to be able to execute such a long-running program without waiting through the
times when the software is sleeping.

Beyond being able to skip over sleeping time and being able to remove the effect of 
your own code's run time, one can envision additional time-related needs like:

* _Simulate different and varying CPU speeds:_ This applies e.g. if one wants to
test whether the control software behaves accurately in extreme situation when
it is starved out of CPU resources by other running processes (e.g. a web
server experiencing a denial-of-service attack).

## Solution design

We began with the intuition that a piece of Go code perceives the passing of time
in a very specific way: by invoking few functions like `time.Sleep` that
can speak to the Go runtime. In other words, if we could override the value returned
by `time.Sleep` with an arbitrary value we would, in principle, be able to
control the running algorithm's perception of time.

At a high-level, this coould be done in two ways: By modifying the Go runtime (and thereby using
a modified Go compiler) or by seeking a way to systematically rewrite the target code
in a way that achieves the same effect. We found a sweet combination of both that 
does not require a modified compiler.

Imagine you are the Go runtime and you are executing some code. You want to
to avoid waiting idle at times when all goroutines are asleep. Your goal is
to find a lazy way of tricking the hosted algorithm, without having to reason
about what it does as a whole.

Notice that you can observe the calls to `time.Now` (to read the current time)
and `time.Sleep` (to fall asleep for a while) which, short of some other
time-related library functions, are the only way the hosted algorithm knows
what time it is. And you control them. What could you do?

Take, for example, this code:

		t0 := time.Now()
		time.Sleep(time.Second)
		println(time.Now().Sub(t0))

To fool it:

* On the first invokation to `time.Now()` we simply return the true time
* Following, we return immediately from the call to `time.Sleep`, thereby saving ourselves waiting for a whole second
* And, finally, to make sure the algorithm cannot distinguish between
the modified and unmodified execution, we arrange that the second
call time `time.Now` returns a value so as to have `time.Now().Sub(t0)` (the
difference in the times returned by the two calls to `time.Now`) equal
`time.Second` (one second).

This technique can easily be generalized to any code, for the case of a single
goroutine. It turns out, and this is a harder exercise, that one can fool an
algorithm that uses a variable number of goroutines. To do this, however, one
needs to modify (the return values of) not just invokations to `time.Now` and 
`time.Sleep`, but also to keep track of goroutines' creation and death events 
as well as channel operation events (like _"message was sent in goroutine A"_
and _"message was received in goroutine B"_).

This can all be done if he have access to the runtime's source and are
willing to change it. However, we chose to take a different, less invasive
route.

### Source-to-source transformer and a runtime inside a runtime

## 

To install,

	$ go get github.com/petar/vitamix

## About

_vitamix_ is authored and maintained by [Petar Maymounkov](http://pdos.csail.mit.edu/~petar/). 
