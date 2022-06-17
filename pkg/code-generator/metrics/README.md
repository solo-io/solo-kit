# Metrics
Tool used primarily for development, to help triage bottlenecks experience during codegen

*This tool was introduced when working on [optimizations to codegen](https://github.com/solo-io/solo-kit/pull/417) and has been preserved for use in future development*

### MeasureElapsed
A convenient way to measure how long an operation took to run:

```go
func longRunningOperation() {
    defer metrics.MeasureElapsed("long-running-operation", time.Now())

    ...
}
```

### IncrementFrequency
A convenient way to count the number of times an operation is run:

```go
func frequentOperation() {
    metrics.IncrementFrequency("frequent-operation")

    ...
}
```

### Flush
A way to flush the aggregate metrics at the end of a run:

```go
func confusingWorkflow() {
    ...

    metrics.Flush(os.Stdout)
}
```