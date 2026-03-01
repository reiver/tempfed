# www/

In here are `http.Handler`.

The directory structure is suggestive of what path they handle.

The `http.Handler` in each package in and under `www/` automatically registers itself with `httpsrv.Mux` at `init()` time.

## Coupling

Although source-code in `www/` can import from anywhere, other parts of this source-code (outside of `www/`) MUST NOT import anything from or under `www/`.

## Package-Names

To drive this point, all `package` names in and under `www/` is `verboten`.
I.e., **forbidden**.

All package names in and under `www/` MUST be `verboten`.
I.e.,:

```golang
package verboten

// ...
```
