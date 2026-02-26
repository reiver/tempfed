# lib/

In here are software libraries, called packages in the Go programming-language (golang).

Source-Code under `lib/` MUST NOT import from any other part of this source-code base!
It can import from other things under `lib/`, and it can import 3rd party packages.

Package-Names for packages under `lib/` should start with a `lib` prefix.
For example: `libdb`, `libping`, `libpong`, `libzip`, etc.
