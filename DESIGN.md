
## Solution design

We began with the intuition that a piece of Go code perceives and interacts
with the passing of time in a very specific and "narrow" way: by invoking a few
standard functions like `time.Now` and `time.Sleep` that speak to the Go
runtime. 

If we could e.g. override the value returned by `time.Sleep` with a
value of our choice we would, in principle, be able to control the running algorithm's
perception of time. We say "in principle" because this isn't the only
crux to the problem. The 

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
* Following, we return immediately from the call to `time.Sleep`, thereby
saving ourselves waiting for a whole second
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

