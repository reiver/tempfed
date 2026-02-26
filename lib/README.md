# lib/

In here are **software libraries**, called **packages** in the Go programming-language (golang).

## Coupling

Source-Code under `lib/` MUST NOT import from any other part of this source-code base!
It can import from other things under `lib/`, and it can import 3rd party packages.

## Naming

Package-Names for packages under `lib/` should start with a `lib` prefix.
For example: `libdb` (in `lib/db/`), `libping` (in `lib/ping/`), `libpong` (in `lib/pon/`), `libzip` (in `lib/zip/`), etc.
