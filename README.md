# run node processes and talk to them

We needed a fast way to render react views on the server. We decided the most efficient way in our case would be to spawn node processes and talk json to them through stdin / stdout / stderr.

```go
flock, err := nodeflock.NewFlock("path/to/module.js", 5)

result, console, err := flock.Call("WhatEver.Static.Call", "hello js")

```