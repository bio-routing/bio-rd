Statusz-like page for Go
========================

This module adds a /debug/status handler to net/http that displays useful information for debug purposes in production.

Basic Usage
-----------

For the basic status page, just include the module.

```go
import (
    _ "github.com/q3k/statusz"
)

func main() {
    http.ListenAndServe("127.0.0.1:6000", nil)
}
```

Adding sections
---------------

To add a section to the status page, call `AddStatusSection` like so:

```go
import (
    statusz "github.com/q3k/statusz"
)


func main() {
    statusz.AddStatusPart("Worker Status", function(ctx context.Context) {
        return fmt.Sprintf("%d workers alive", workerCount)
    })
    http.ListenAndServe("127.0.0.1:6000", nil)
}
```

For custom section templates, call `AddStatusPart`, which accepts a http/template fragment that will be rendered on the result of the part render function.
