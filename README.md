# lukasmalkmus/rpi_exporter

> A Raspberry Pi CPU temperature exporter. - by **[Lukas Malkmus]** (fork by [cedi])

[![go report_badge]][report]
[![release_badge]][release page]
[![license_badge]][license]

![go_releaser_build_badge]
![docker_build_badge]

---

## Table of Contents

1. [Introduction](#introduction)
2. [Usage](#usage)
3. [Contributing](#contributing)
4. [License](#license)

### Introduction

The *rpi_exporter* is a simple server that scrapes the Raspberry Pi's CPU
temperature and exports it via HTTP for Prometheus consumption.

### Usage

#### Installation

The easiest way to run the *rpi_exporter* is by grabbing the latest binary from
the [release page].

Do not forget to run *rpi_exporter* using user in `video` group to get GPU
details from RPi.

##### Building from source

This project uses [go mod] for vendoring.

```bash
git clone https://github.com/lukasmalkmus/rpi_exporter.git
cd rpi_exporter
make build
```

#### Using the application

```bash
./rpi_exporter [flags]
```

Help on flags:

```bash
./rpi_exporter --help
```

### Contributing

Feel free to submit PRs or to fill Issues. Every kind of help is appreciated.

### License

© Lukas Malkmus, 2019

Distributed under Apache License (`Apache License, Version 2.0`).

See [LICENSE](LICENSE) for more information.

<!-- Links -->
[go mod]: https://golang.org/cmd/go/#hdr-Module_maintenance
[Lukas Malkmus]: https://github.com/lukasmalkmus
[cedi]: https://github.com/cedi

<!-- Badges -->
[go report_badge]: https://goreportcard.com/badge/github.com/cedi/rpi_exporter
[report]: https://goreportcard.com/report/github.com/cedi/rpi_exporter
[release page]: https://github.com/cedi/rpi_exporter/releases
[release_badge]: https://img.shields.io/github/release/cedi/rpi_exporter.svg
[license]: https://opensource.org/licenses/Apache-2.0
[license_badge]: https://img.shields.io/badge/license-Apache-blue.svg
[go_releaser_build_badge]: https://github.com/cedi/rpi_exporter/actions/workflows/ro_releaser/badge.svg
[docker_build_badge]: https://github.com/cedi/rpi_exporter/actions/workflows/docker_build/badge.svg
