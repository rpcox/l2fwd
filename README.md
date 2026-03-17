# l2fwd

Needed a swiss army knife for work. Wrote it a while back and had to remove all the 'syscall' references

Can compile on Mac w/ build file. Linux direct compile still has an issue to resolve

    # runtime
    /usr/local/go/src/runtime/traceexp.go:22:6: unsafeTraceExpWriter redeclared in this block
	/usr/local/go/src/runtime/tracebuf.go:220:6: other declaration of unsafeTraceExpWriter
    /usr/local/go/src/runtime/traceexp.go:32:40: too many arguments in call to w.traceWriter.refill
	have (traceExperiment)
	want ()
    /usr/local/go/src/runtime/traceexp.go:50:4: undefined: traceEv
