# git-semver

Manage semantic version tags in a git repository.

Follows the command-line interface and configuration syntax of https://github.com/markchalloner/git-semver but this
version is implemented using https://github.com/Masterminds/semver to parse and manipulate semantic versions.

## Installation

```bash
go install github.com/ray1729/git-semver@v0.2.1
```

## Usage

```bash
git-semver help
```

## Configuration

Configuration is read from the first of these paths that is found: `$PWD/.git-semver`, `$XDG_CONFIG_HOME/git-semver`, `$HOME/.config/git-semver`, `$HOME/.git-semver/config`.

Configuration format is one `key=value` pair per line, lines beginning with `#` are ignored. The following keys are
understood:

* `VERSION_PREFIX` specify a prefix string for the created version tags
* `GIT_SIGN` boolean value indicating whether or not to sign tags

## License

MIT License

Copyright 2023 Raymond Miller <ray@1729.org.uk>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
