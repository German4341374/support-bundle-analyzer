# JVM thread dump analysis

Capture multiple thread dumps several seconds apart when possible. One snapshot can show a state but not whether it persists.

The JVM analyzer counts thread states, identifies deadlock markers, repeated stacks, blocked threads, long GC pauses, frequent Full GC, OutOfMemoryError evidence, and startup metadata. A large heap dump is classified but not parsed; use Eclipse MAT or another specialist tool in a controlled environment.

Compare repeated stacks across captures and correlate them with GC and application logs. Many WAITING threads can be normal. BLOCKED growth on the same monitor is more actionable but still requires code and timing context.

Thread names, arguments, system properties, and stack values can expose credentials or customer identifiers. Review and redact before sharing.
