# cfg/

In here are **configurations**.

Source-Code in `cfg/` MUST NOT import from any other part of this source-code base!
It can import 3rd party packages.

## Environment-Variables

Configurations comes from environment-variables.

So, for example, if this software was compiled to the executable file: `tempfed`.
Then running `tempfed` with configuration environment-varaibles would might something like:

```bash
MSTDN_HOSTS="fedi.buzz, relay.example" ./tempfed
```

The `MSTDN_HOSTS` environment-varaible would get returned to the source-code by the `cfg.MstdnHosts()` function.
In the case of our example, `cfg.MstdnHosts()` would return:

```golang
[]string{"fedi.buzz", "relay.example"}
```
